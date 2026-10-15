package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/pnz1990/krombat/backend/internal/game"
	"github.com/pnz1990/krombat/backend/internal/k8s"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

const ownerLabel = "krombat.io/owner"

var validDNSLabel = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

type Handler struct {
	client     dynamic.Interface
	auth       *sessionAuth
	cel        *celEvaluator
	loginLimit *rateLimiter
	celLimit   *rateLimiter
}

func New(client dynamic.Interface, secret string) (*Handler, error) {
	auth, err := newSessionAuth(secret)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("Kubernetes dynamic client is required")
	}
	evaluator, err := newCELEvaluator()
	if err != nil {
		return nil, fmt.Errorf("initialize upstream kro CEL environment: %w", err)
	}
	return &Handler{
		client: client, auth: auth, cel: evaluator,
		loginLimit: newRateLimiter(time.Second),
		celLimit:   newRateLimiter(100 * time.Millisecond),
	}, nil
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/v1/auth/login", h.loginLimit.Wrap(h.Login, remoteIP))
	mux.HandleFunc("GET /api/v1/auth/me", h.auth.require(h.Me))
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)
	mux.HandleFunc("GET /api/v1/dungeons", h.auth.require(h.ListDungeons))
	mux.HandleFunc("POST /api/v1/dungeons", h.auth.require(h.CreateDungeon))
	mux.HandleFunc("GET /api/v1/dungeons/{namespace}/{name}", h.auth.require(h.GetDungeon))
	mux.HandleFunc("DELETE /api/v1/dungeons/{namespace}/{name}", h.auth.require(h.DeleteDungeon))
	mux.HandleFunc("POST /api/v1/dungeons/{namespace}/{name}/attacks", h.auth.require(h.CreateAttack))
	mux.HandleFunc("POST /api/v1/dungeons/{namespace}/{name}/actions", h.auth.require(h.CreateAction))
	mux.HandleFunc("GET /api/v1/dungeons/{namespace}/{name}/resources", h.auth.require(h.GetResources))
	mux.HandleFunc("POST /api/v1/dungeons/{namespace}/{name}/cel-eval",
		h.auth.require(h.celLimit.Wrap(h.CELEval, remoteIP)))
	return AccessLog(mux)
}

type CreateDungeonReq struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Monsters   int64  `json:"monsters"`
	Difficulty string `json:"difficulty"`
	HeroClass  string `json:"heroClass"`
}

func (h *Handler) CreateDungeon(w http.ResponseWriter, r *http.Request) {
	var req CreateDungeonReq
	if !decodeBody(w, r, &req) {
		return
	}
	if !validDNSLabel.MatchString(req.Name) {
		writeError(w, "name must be a lowercase DNS label, at most 63 characters", http.StatusBadRequest)
		return
	}
	if req.Namespace != "" && req.Namespace != k8s.GameNamespace {
		writeError(w, "namespace must be rpg-demo", http.StatusBadRequest)
		return
	}
	spec, err := game.NewSpec(req.Monsters, req.Difficulty, req.HeroClass)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	specMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&spec)
	if err != nil {
		writeError(w, "could not initialize game state", http.StatusInternalServerError)
		return
	}
	dungeon := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": k8s.DungeonGVR.GroupVersion().String(),
		"kind":       "Dungeon",
		"metadata": map[string]interface{}{
			"name": req.Name, "namespace": k8s.GameNamespace,
			"labels": map[string]interface{}{ownerLabel: demoLogin},
		},
		"spec": specMap,
	}}
	created, err := h.client.Resource(k8s.DungeonGVR).Namespace(k8s.GameNamespace).
		Create(r.Context(), dungeon, metav1.CreateOptions{})
	if err != nil {
		writeK8sError(w, "create dungeon", err)
		return
	}
	writeJSON(w, http.StatusCreated, created.Object)
}

func (h *Handler) ListDungeons(w http.ResponseWriter, r *http.Request) {
	if ns := r.URL.Query().Get("namespace"); ns != "" && ns != k8s.GameNamespace {
		writeError(w, "namespace must be rpg-demo", http.StatusBadRequest)
		return
	}
	list, err := h.client.Resource(k8s.DungeonGVR).Namespace(k8s.GameNamespace).
		List(r.Context(), metav1.ListOptions{LabelSelector: ownerLabel + "=" + demoLogin})
	if err != nil {
		writeK8sError(w, "list dungeons", err)
		return
	}
	items := make([]map[string]interface{}, 0, len(list.Items))
	for _, dungeon := range list.Items {
		if dungeon.GetNamespace() == k8s.GameNamespace &&
			dungeon.GetLabels()[ownerLabel] == demoLogin && dungeon.GetDeletionTimestamp() == nil {
			items = append(items, dungeon.Object)
		}
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) ownedDungeon(w http.ResponseWriter, r *http.Request) *unstructured.Unstructured {
	if r.PathValue("namespace") != k8s.GameNamespace {
		writeError(w, "namespace must be rpg-demo", http.StatusBadRequest)
		return nil
	}
	name := r.PathValue("name")
	if !validDNSLabel.MatchString(name) {
		writeError(w, "invalid dungeon name", http.StatusBadRequest)
		return nil
	}
	dungeon, err := h.client.Resource(k8s.DungeonGVR).Namespace(k8s.GameNamespace).
		Get(r.Context(), name, metav1.GetOptions{})
	if err != nil {
		writeK8sError(w, "get dungeon", err)
		return nil
	}
	if dungeon.GetLabels()[ownerLabel] != demoLogin || dungeon.GetNamespace() != k8s.GameNamespace {
		writeError(w, "dungeon not found", http.StatusNotFound)
		return nil
	}
	return dungeon
}

func (h *Handler) GetDungeon(w http.ResponseWriter, r *http.Request) {
	if dungeon := h.ownedDungeon(w, r); dungeon != nil {
		writeJSON(w, http.StatusOK, dungeon.Object)
	}
}

func (h *Handler) DeleteDungeon(w http.ResponseWriter, r *http.Request) {
	dungeon := h.ownedDungeon(w, r)
	if dungeon == nil {
		return
	}
	preconditions := &metav1.Preconditions{}
	if uid := dungeon.GetUID(); uid != "" {
		preconditions.UID = &uid
	}
	if version := dungeon.GetResourceVersion(); version != "" {
		preconditions.ResourceVersion = &version
	}
	policy := metav1.DeletePropagationBackground
	if err := h.client.Resource(k8s.DungeonGVR).Namespace(k8s.GameNamespace).
		Delete(r.Context(), dungeon.GetName(), metav1.DeleteOptions{
			Preconditions: preconditions, PropagationPolicy: &policy,
		}); err != nil {
		writeK8sError(w, "delete dungeon", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type CreateAttackReq struct {
	Target string          `json:"target"`
	Seq    *int64          `json:"seq"`
	Damage json.RawMessage `json:"damage,omitempty"`
}

func (h *Handler) CreateAttack(w http.ResponseWriter, r *http.Request) {
	var req CreateAttackReq
	if !decodeBody(w, r, &req) {
		return
	}
	h.applyTurn(w, r, req.Seq, func(spec game.Spec) (game.Spec, error) {
		if req.Target == "open-treasure" {
			return spec.Action(req.Target)
		}
		return spec.Attack(r.PathValue("name"), req.Target)
	})
}

func (h *Handler) CreateAction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Action string `json:"action"`
		Seq    *int64 `json:"seq"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	h.applyTurn(w, r, req.Seq, func(spec game.Spec) (game.Spec, error) {
		return spec.Action(req.Action)
	})
}

func (h *Handler) applyTurn(w http.ResponseWriter, r *http.Request, seq *int64,
	apply func(game.Spec) (game.Spec, error),
) {
	if seq == nil || *seq < 0 {
		writeError(w, "seq is required and must be a nonnegative integer", http.StatusBadRequest)
		return
	}
	dungeon := h.ownedDungeon(w, r)
	if dungeon == nil {
		return
	}
	if dungeon.GetDeletionTimestamp() != nil {
		writeError(w, "dungeon is being deleted", http.StatusConflict)
		return
	}
	spec, err := dungeonSpec(dungeon)
	if err != nil {
		slog.Error("invalid dungeon state", "dungeon", dungeon.GetName(), "error", err)
		writeError(w, "dungeon has invalid spec.game state", http.StatusInternalServerError)
		return
	}
	if *seq != spec.AttackSeq {
		writeError(w, "stale seq; refresh the dungeon and retry", http.StatusConflict)
		return
	}
	next, err := apply(spec)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, game.ErrSequenceExhausted) {
			status = http.StatusConflict
		}
		writeError(w, err.Error(), status)
		return
	}
	if dungeon.GetResourceVersion() == "" {
		writeError(w, "dungeon is missing resourceVersion", http.StatusInternalServerError)
		return
	}
	nextMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&next)
	if err != nil {
		writeError(w, "could not encode game state", http.StatusInternalServerError)
		return
	}
	updated := dungeon.DeepCopy()
	updated.Object["spec"] = nextMap
	// Keep the fetched resourceVersion. A competing turn or kro reconcile must
	// produce 409, never a blind retry that could apply the same turn twice.
	result, err := h.client.Resource(k8s.DungeonGVR).Namespace(k8s.GameNamespace).
		Update(r.Context(), updated, metav1.UpdateOptions{})
	if err != nil {
		writeK8sError(w, "update dungeon turn", err)
		return
	}
	writeJSON(w, http.StatusOK, result.Object)
}

func dungeonSpec(dungeon *unstructured.Unstructured) (game.Spec, error) {
	specMap, found, err := unstructured.NestedMap(dungeon.Object, "spec")
	if err != nil || !found {
		return game.Spec{}, errors.New("missing spec")
	}
	state, found, err := unstructured.NestedMap(specMap, "game")
	if err != nil || !found {
		return game.Spec{}, errors.New("missing spec.game")
	}
	for _, field := range []string{"heroHP", "bossHP", "monsterHP", "treasureOpened"} {
		if _, ok := state[field]; !ok {
			return game.Spec{}, fmt.Errorf("missing spec.game.%s", field)
		}
	}
	for _, field := range []string{"monsters", "difficulty", "heroClass", "attackSeq"} {
		if _, ok := specMap[field]; !ok {
			return game.Spec{}, fmt.Errorf("missing spec.%s", field)
		}
	}
	var spec game.Spec
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(specMap, &spec); err != nil {
		return game.Spec{}, err
	}
	return spec, spec.Validate()
}

func decodeBody(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(target)
	if err == nil {
		var extra interface{}
		if extraErr := decoder.Decode(&extra); extraErr != io.EOF {
			err = extraErr
			if err == nil {
				err = errors.New("expected one JSON object")
			}
		}
	}
	if err == nil {
		return true
	}
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeError(w, "request body too large", http.StatusRequestEntityTooLarge)
	} else {
		writeError(w, "invalid request body: expected one JSON object with supported fields", http.StatusBadRequest)
	}
	return false
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("encode API response", "error", err)
	}
}

func writeError(w http.ResponseWriter, message string, status int) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeK8sError(w http.ResponseWriter, operation string, err error) {
	switch {
	case apierrors.IsNotFound(err):
		writeError(w, "dungeon not found", http.StatusNotFound)
	case apierrors.IsAlreadyExists(err):
		writeError(w, "dungeon already exists", http.StatusConflict)
	case apierrors.IsConflict(err):
		writeError(w, "dungeon changed; refresh and retry with the current seq", http.StatusConflict)
	default:
		slog.Error("Kubernetes request failed", "operation", operation, "error", err)
		writeError(w, "Kubernetes request failed; retry shortly", http.StatusServiceUnavailable)
	}
}
