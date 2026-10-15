package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pnz1990/krombat/backend/internal/game"
	"github.com/pnz1990/krombat/backend/internal/k8s"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	clienttesting "k8s.io/client-go/testing"
)

const (
	testSecret = "local-unit-test-secret-32-bytes-long"
	testPath   = "/api/v1/dungeons/rpg-demo/demo-room"
)

func newTestAPI(t *testing.T, objects ...runtime.Object) (*Handler, *fake.FakeDynamicClient, http.Handler) {
	t.Helper()
	kinds := map[schema.GroupVersionResource]string{k8s.DungeonGVR: "DungeonList"}
	listKinds := map[string]string{
		"heroes": "HeroList", "monsters": "MonsterList", "bosses": "BossList",
		"treasures": "TreasureList", "loots": "LootList",
		"configmaps": "ConfigMapList", "secrets": "SecretList",
	}
	for _, resource := range inspectableResources {
		kinds[resource] = listKinds[resource.Resource]
	}
	client := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), kinds)
	for _, object := range objects {
		obj := object.(*unstructured.Unstructured)
		found := false
		for gvr, listKind := range kinds {
			if obj.GetKind()+"List" != listKind || obj.GetAPIVersion() != gvr.GroupVersion().String() {
				continue
			}
			if err := client.Tracker().Create(gvr, obj, obj.GetNamespace()); err != nil {
				t.Fatal(err)
			}
			found = true
			break
		}
		if !found {
			t.Fatalf("unregistered fixture kind %s", obj.GetKind())
		}
	}
	h, err := New(client, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	h.celLimit.interval = 0
	return h, client, h.Routes()
}

func testDungeon(t *testing.T, difficulty string) *unstructured.Unstructured {
	t.Helper()
	spec, err := game.NewSpec(2, difficulty, "warrior")
	if err != nil {
		t.Fatal(err)
	}
	specMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&spec)
	if err != nil {
		t.Fatal(err)
	}
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "game.k8s.example/v1alpha1", "kind": "Dungeon",
		"metadata": map[string]interface{}{
			"name": "demo-room", "namespace": k8s.GameNamespace,
			"uid": "dungeon-uid", "resourceVersion": "7",
			"labels": map[string]interface{}{ownerLabel: demoLogin},
		},
		"spec": specMap,
		"status": map[string]interface{}{
			"bossPhase": "phase1",
			"game": map[string]interface{}{
				"heroHP": int64(0), "bossHP": int64(0),
				"monsterHP": []interface{}{int64(0), int64(0)},
			},
		},
	}}
}

func request(api http.Handler, method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "127.0.0.1:12345"
	if cookie != nil {
		req.AddCookie(cookie)
	}
	recorder := httptest.NewRecorder()
	api.ServeHTTP(recorder, req)
	return recorder
}

func login(t *testing.T, api http.Handler) *http.Cookie {
	t.Helper()
	result := request(api, http.MethodPost, "/api/v1/auth/login", "", nil)
	if result.Code != http.StatusOK {
		t.Fatalf("login: %d %s", result.Code, result.Body.String())
	}
	cookies := result.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != sessionCookieName {
		t.Fatalf("expected session cookie, got %v", cookies)
	}
	return cookies[0]
}

func responseDungeon(t *testing.T, response *httptest.ResponseRecorder) *unstructured.Unstructured {
	t.Helper()
	var dungeon unstructured.Unstructured
	if err := dungeon.UnmarshalJSON(response.Body.Bytes()); err != nil {
		t.Fatalf("decode dungeon response %s: %v", response.Body.String(), err)
	}
	return &dungeon
}

func responseSpec(t *testing.T, response *httptest.ResponseRecorder) game.Spec {
	t.Helper()
	spec, err := dungeonSpec(responseDungeon(t, response))
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func setSpec(t *testing.T, dungeon *unstructured.Unstructured, mutate func(*game.Spec)) {
	t.Helper()
	spec, err := dungeonSpec(dungeon)
	if err != nil {
		t.Fatal(err)
	}
	mutate(&spec)
	specMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&spec)
	if err != nil {
		t.Fatal(err)
	}
	dungeon.Object["spec"] = specMap
}

func TestCreateDungeon(t *testing.T) {
	_, client, api := newTestAPI(t)
	cookie := login(t, api)
	response := request(api, http.MethodPost, "/api/v1/dungeons",
		`{"name":"new-demo","monsters":3,"difficulty":"hard"}`, cookie)
	if response.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", response.Code, response.Body.String())
	}
	dungeon := responseDungeon(t, response)
	spec := responseSpec(t, response)
	if dungeon.GetName() != "new-demo" || dungeon.GetNamespace() != k8s.GameNamespace ||
		dungeon.GetLabels()[ownerLabel] != demoLogin || spec.HeroClass != "warrior" ||
		spec.Game.HeroHP != 200 || spec.Game.BossHP != 800 || spec.AttackSeq != 0 ||
		!reflect.DeepEqual(spec.Game.MonsterHP, []int64{80, 80, 80}) {
		t.Fatalf("incorrect initialized dungeon: %v", dungeon.Object)
	}
	actions := client.Actions()
	if len(actions) != 1 || actions[0].GetNamespace() != k8s.GameNamespace || actions[0].GetVerb() != "create" {
		t.Fatalf("unexpected Kubernetes operations: %v", actions)
	}
	response = request(api, http.MethodPost, "/api/v1/dungeons",
		`{"name":"new-demo","monsters":3,"difficulty":"hard"}`, cookie)
	if response.Code != http.StatusConflict {
		t.Fatalf("duplicate create: %d %s", response.Code, response.Body.String())
	}
}

func TestCreateDungeonValidation(t *testing.T) {
	for _, body := range []string{
		`{`, `null`, `{}`, `{"name":"BAD","monsters":1,"difficulty":"easy"}`,
		`{"name":"-bad","monsters":1,"difficulty":"easy"}`,
		`{"name":"bad-","monsters":1,"difficulty":"easy"}`,
		`{"name":"a.b","monsters":1,"difficulty":"easy"}`,
		`{"name":"` + strings.Repeat("a", 64) + `","monsters":1,"difficulty":"easy"}`,
		`{"name":"ok","monsters":0,"difficulty":"easy"}`,
		`{"name":"ok","monsters":11,"difficulty":"easy"}`,
		`{"name":"ok","monsters":1.5,"difficulty":"easy"}`,
		`{"name":"ok","monsters":1,"difficulty":"impossible"}`,
		`{"name":"ok","monsters":1,"difficulty":"easy","heroClass":"mage"}`,
		`{"name":"ok","monsters":1,"difficulty":"easy","heroClass":"rogue"}`,
		`{"name":"ok","monsters":1,"difficulty":"easy","namespace":"default"}`,
		`{"name":"ok","monsters":1,"difficulty":"easy","namespace":"kube-system"}`,
		`{"name":"ok","monsters":1,"difficulty":"easy","game":{"heroHP":9999}}`,
		`{"name":"ok","monsters":1,"difficulty":"easy"} {}`,
	} {
		t.Run(body, func(t *testing.T) {
			_, client, api := newTestAPI(t)
			response := request(api, http.MethodPost, "/api/v1/dungeons", body, login(t, api))
			if response.Code != http.StatusBadRequest {
				t.Fatalf("got %d: %s", response.Code, response.Body.String())
			}
			if len(client.Actions()) != 0 {
				t.Fatal("invalid creation touched Kubernetes")
			}
		})
	}
	_, _, api := newTestAPI(t)
	response := request(api, http.MethodPost, "/api/v1/dungeons",
		`{"name":"`+strings.Repeat("a", 9000)+`"}`, login(t, api))
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized body: %d %s", response.Code, response.Body.String())
	}
}

func TestListDungeonsReturnsOwnedRawObjects(t *testing.T) {
	own := testDungeon(t, "easy")
	other := own.DeepCopy()
	other.SetName("other")
	other.SetLabels(map[string]string{ownerLabel: "someone-else"})
	unlabelled := own.DeepCopy()
	unlabelled.SetName("unlabelled")
	unlabelled.SetLabels(nil)
	outside := own.DeepCopy()
	outside.SetNamespace("rpg-system")
	deleting := own.DeepCopy()
	deleting.SetName("deleting")
	now := metav1.Now()
	deleting.SetDeletionTimestamp(&now)
	_, client, api := newTestAPI(t, own, other, unlabelled, outside, deleting)
	response := request(api, http.MethodGet, "/api/v1/dungeons", "", login(t, api))
	if response.Code != http.StatusOK {
		t.Fatalf("list: %d %s", response.Code, response.Body.String())
	}
	var items []unstructured.Unstructured
	if err := json.Unmarshal(response.Body.Bytes(), &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].GetName() != own.GetName() ||
		items[0].Object["spec"] == nil || items[0].Object["kind"] != "Dungeon" {
		t.Fatalf("list did not return raw owned Dungeon array: %s", response.Body.String())
	}
	action := client.Actions()[0].(clienttesting.ListAction)
	if action.GetNamespace() != k8s.GameNamespace ||
		action.GetListRestrictions().Labels.String() != ownerLabel+"="+demoLogin {
		t.Fatalf("list not namespace/owner scoped: %+v", action)
	}
}

func TestReadDeleteOwnershipAndPreconditions(t *testing.T) {
	for _, owner := range []string{"", "someone-else"} {
		t.Run("hidden-"+owner, func(t *testing.T) {
			dungeon := testDungeon(t, "easy")
			dungeon.SetLabels(map[string]string{ownerLabel: owner})
			_, client, api := newTestAPI(t, dungeon)
			cookie := login(t, api)
			for _, method := range []string{http.MethodGet, http.MethodDelete} {
				response := request(api, method, testPath, "", cookie)
				if response.Code != http.StatusNotFound {
					t.Fatalf("%s unowned: %d", method, response.Code)
				}
			}
			for _, action := range client.Actions() {
				if action.GetVerb() != "get" {
					t.Fatalf("unowned dungeon modified: %v", action)
				}
			}
		})
	}
	dungeon := testDungeon(t, "easy")
	_, client, api := newTestAPI(t, dungeon)
	cookie := login(t, api)
	if response := request(api, http.MethodGet, testPath, "", cookie); response.Code != http.StatusOK {
		t.Fatalf("get: %d", response.Code)
	}
	client.PrependReactor("delete", "dungeons", func(action clienttesting.Action) (bool, runtime.Object, error) {
		opts := action.(clienttesting.DeleteAction).GetDeleteOptions()
		if opts.Preconditions == nil || opts.Preconditions.UID == nil ||
			*opts.Preconditions.UID != dungeon.GetUID() || opts.Preconditions.ResourceVersion == nil ||
			*opts.Preconditions.ResourceVersion != "7" {
			t.Errorf("delete must use identity/version preconditions: %+v", opts)
		}
		return false, nil, nil
	})
	if response := request(api, http.MethodDelete, testPath, "", cookie); response.Code != http.StatusNoContent {
		t.Fatalf("delete: %d %s", response.Code, response.Body.String())
	}
	if response := request(api, http.MethodGet, testPath, "", cookie); response.Code != http.StatusNotFound {
		t.Fatalf("get deleted: %d", response.Code)
	}
}

func TestNamespaceAndNameValidation(t *testing.T) {
	_, client, api := newTestAPI(t)
	cookie := login(t, api)
	for _, path := range []string{
		"/api/v1/dungeons?namespace=default",
		"/api/v1/dungeons/default/demo-room",
		"/api/v1/dungeons/kube-system/demo-room/resources",
		"/api/v1/dungeons/rpg-demo/BAD",
	} {
		response := request(api, http.MethodGet, path, "", cookie)
		if response.Code != http.StatusBadRequest {
			t.Errorf("%s: %d %s", path, response.Code, response.Body.String())
		}
	}
	if len(client.Actions()) != 0 {
		t.Fatal("invalid paths touched Kubernetes")
	}
}

func TestAttackIgnoresClientDamageAndStaleStatus(t *testing.T) {
	for _, damage := range []string{"999999", "-20", `"not authoritative"`, "null"} {
		t.Run(damage, func(t *testing.T) {
			dungeon := testDungeon(t, "normal")
			_, client, api := newTestAPI(t, dungeon)
			response := request(api, http.MethodPost, testPath+"/attacks",
				`{"target":"demo-room-monster-0","seq":0,"damage":`+damage+`}`, login(t, api))
			if response.Code != http.StatusOK {
				t.Fatalf("attack: %d %s", response.Code, response.Body.String())
			}
			spec := responseSpec(t, response)
			if spec.Game.HeroHP != 197 || spec.Game.BossHP != 400 || spec.AttackSeq != 1 ||
				!reflect.DeepEqual(spec.Game.MonsterHP, []int64{15, 50}) {
				t.Fatalf("client damage/status influenced turn: %+v", spec)
			}
			update := client.Actions()[1].(clienttesting.UpdateAction)
			updated := update.GetObject().(*unstructured.Unstructured)
			if updated.GetResourceVersion() != "7" || update.GetSubresource() != "" {
				t.Fatalf("update did not preserve resourceVersion: %+v", update)
			}
			for _, action := range client.Actions() {
				if action.GetVerb() == "patch" || action.GetResource() != k8s.DungeonGVR {
					t.Fatalf("unexpected legacy write: %v", action)
				}
			}
		})
	}
}

func TestAttackErrorsDoNotMutate(t *testing.T) {
	for _, tt := range []struct {
		name   string
		body   string
		mutate func(*game.Spec)
		status int
		error  string
	}{
		{"missing seq", `{"target":"demo-room-monster-0"}`, nil, 400, "seq is required"},
		{"null seq", `{"target":"demo-room-monster-0","seq":null}`, nil, 400, "seq is required"},
		{"negative seq", `{"target":"demo-room-monster-0","seq":-1}`, nil, 400, "seq is required"},
		{"fractional seq", `{"target":"demo-room-monster-0","seq":0.5}`, nil, 400, "invalid request"},
		{"stale seq", `{"target":"demo-room-monster-0","seq":0}`, func(s *game.Spec) { s.AttackSeq = 1 }, 409, "stale seq"},
		{"future seq", `{"target":"demo-room-monster-0","seq":1}`, nil, 409, "stale seq"},
		{"locked boss", `{"target":"demo-room-boss","seq":0}`, nil, 400, "boss is locked"},
		{"wrong target", `{"target":"somebody-monster-0","seq":0}`, nil, 400, "exact monster name"},
		{"unknown target", `{"target":"demo-room-monster-2","seq":0}`, nil, 400, "exact monster name"},
		{"dead target", `{"target":"demo-room-monster-0","seq":0}`, func(s *game.Spec) { s.Game.MonsterHP[0] = 0 }, 400, "already defeated"},
		{"defeat", `{"target":"demo-room-monster-0","seq":0}`, func(s *game.Spec) { s.Game.HeroHP = 0 }, 400, "hero is defeated"},
		{"victory", `{"target":"demo-room-boss","seq":0}`, func(s *game.Spec) { s.Game.BossHP = 0 }, 400, "combat is complete"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dungeon := testDungeon(t, "easy")
			if tt.mutate != nil {
				setSpec(t, dungeon, tt.mutate)
			}
			_, client, api := newTestAPI(t, dungeon)
			response := request(api, http.MethodPost, testPath+"/attacks", tt.body, login(t, api))
			if response.Code != tt.status || !strings.Contains(response.Body.String(), tt.error) {
				t.Fatalf("got %d %s", response.Code, response.Body.String())
			}
			for _, action := range client.Actions() {
				if action.GetVerb() != "get" {
					t.Fatalf("invalid attack mutated Kubernetes: %v", action)
				}
			}
		})
	}
}

func TestActionUsesSharedAttackSequence(t *testing.T) {
	dungeon := testDungeon(t, "easy")
	setSpec(t, dungeon, func(s *game.Spec) {
		s.AttackSeq = 9
		s.Game.BossHP = 0
		s.Game.MonsterHP = []int64{0, 0}
	})
	_, _, api := newTestAPI(t, dungeon)
	cookie := login(t, api)
	response := request(api, http.MethodPost, testPath+"/attacks",
		`{"target":"open-treasure","seq":9}`, cookie)
	if response.Code != http.StatusOK {
		t.Fatalf("open treasure: %d %s", response.Code, response.Body.String())
	}
	spec := responseSpec(t, response)
	if spec.AttackSeq != 10 || spec.Game.TreasureOpened != 1 {
		t.Fatalf("action did not consume the combat sequence: %+v", spec)
	}
	for _, endpoint := range []string{"actions", "attacks"} {
		body := `{"action":"open-treasure","seq":9}`
		if endpoint == "attacks" {
			body = `{"target":"demo-room-boss","seq":9}`
		}
		if response := request(api, http.MethodPost, testPath+"/"+endpoint, body, cookie); response.Code != http.StatusConflict {
			t.Fatalf("repeated %s sequence: %d %s", endpoint, response.Code, response.Body.String())
		}
	}
	response = request(api, http.MethodPost, testPath+"/actions",
		`{"action":"open-treasure","seq":10}`, cookie)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "already open") {
		t.Fatalf("repeated treasure: %d %s", response.Code, response.Body.String())
	}
}

func TestTurnResourceVersionConflictIsNotRetried(t *testing.T) {
	for _, endpoint := range []string{"attacks", "actions"} {
		t.Run(endpoint, func(t *testing.T) {
			dungeon := testDungeon(t, "easy")
			body := `{"target":"demo-room-monster-0","seq":0}`
			if endpoint == "actions" {
				body = `{"action":"open-treasure","seq":0}`
				setSpec(t, dungeon, func(s *game.Spec) { s.Game.BossHP = 0 })
			}
			_, client, api := newTestAPI(t, dungeon)
			updates := 0
			client.PrependReactor("update", "dungeons", func(action clienttesting.Action) (bool, runtime.Object, error) {
				updates++
				return true, nil, apierrors.NewConflict(k8s.DungeonGVR.GroupResource(), "demo-room", errors.New("concurrent reconciliation"))
			})
			response := request(api, http.MethodPost, testPath+"/"+endpoint, body, login(t, api))
			if response.Code != http.StatusConflict || updates != 1 {
				t.Fatalf("resourceVersion race must return one 409: %d, %d updates, %s", response.Code, updates, response.Body.String())
			}
		})
	}
}

type barrierClient struct {
	dynamic.Interface
	mu    sync.Mutex
	reads int
	gate  chan struct{}
}

func (c *barrierClient) Resource(gvr schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return &barrierNamespace{NamespaceableResourceInterface: c.Interface.Resource(gvr), barrier: c}
}

type barrierNamespace struct {
	dynamic.NamespaceableResourceInterface
	barrier *barrierClient
}

func (n *barrierNamespace) Namespace(namespace string) dynamic.ResourceInterface {
	return &barrierResource{ResourceInterface: n.NamespaceableResourceInterface.Namespace(namespace), barrier: n.barrier}
}

type barrierResource struct {
	dynamic.ResourceInterface
	barrier *barrierClient
}

func (r *barrierResource) Get(ctx context.Context, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	obj, err := r.ResourceInterface.Get(ctx, name, options, subresources...)
	if err != nil {
		return nil, err
	}
	r.barrier.mu.Lock()
	r.barrier.reads++
	if r.barrier.reads == 2 {
		close(r.barrier.gate)
	}
	r.barrier.mu.Unlock()
	select {
	case <-r.barrier.gate:
		return obj, nil
	case <-time.After(5 * time.Second):
		return nil, errors.New("concurrent read barrier timed out")
	}
}

func TestSimultaneousTurnsApplyExactlyOnce(t *testing.T) {
	for _, endpoint := range []string{"attacks", "actions"} {
		t.Run(endpoint, func(t *testing.T) {
			dungeon := testDungeon(t, "normal")
			body := `{"target":"demo-room-monster-0","seq":0}`
			if endpoint == "actions" {
				body = `{"action":"open-treasure","seq":0}`
				setSpec(t, dungeon, func(s *game.Spec) { s.Game.BossHP = 0 })
			}
			h, client, api := newTestAPI(t, dungeon)
			h.client = &barrierClient{Interface: client, gate: make(chan struct{})}
			client.PrependReactor("update", "dungeons", func(action clienttesting.Action) (bool, runtime.Object, error) {
				proposed := action.(clienttesting.UpdateAction).GetObject().(*unstructured.Unstructured)
				current, err := client.Tracker().Get(k8s.DungeonGVR, k8s.GameNamespace, "demo-room")
				if err != nil {
					return true, nil, err
				}
				if proposed.GetResourceVersion() != current.(*unstructured.Unstructured).GetResourceVersion() {
					return true, nil, apierrors.NewConflict(k8s.DungeonGVR.GroupResource(), "demo-room", errors.New("stale resourceVersion"))
				}
				proposed = proposed.DeepCopy()
				proposed.SetResourceVersion("8")
				err = client.Tracker().Update(k8s.DungeonGVR, proposed, k8s.GameNamespace)
				return true, proposed, err
			})
			cookie := login(t, api)
			var wg sync.WaitGroup
			responses := make(chan int, 2)
			for range 2 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					responses <- request(api, http.MethodPost, testPath+"/"+endpoint, body, cookie).Code
				}()
			}
			wg.Wait()
			close(responses)
			statuses := map[int]int{}
			for status := range responses {
				statuses[status]++
			}
			if statuses[http.StatusOK] != 1 || statuses[http.StatusConflict] != 1 {
				t.Fatalf("expected one committed turn and one 409: %v", statuses)
			}
			stored, err := client.Tracker().Get(k8s.DungeonGVR, k8s.GameNamespace, "demo-room")
			if err != nil {
				t.Fatal(err)
			}
			spec, err := dungeonSpec(stored.(*unstructured.Unstructured))
			if err != nil || spec.AttackSeq != 1 {
				t.Fatalf("turn applied more than once: %+v, %v", spec, err)
			}
		})
	}
}

func TestInvalidCanonicalStateAndMissingVersion(t *testing.T) {
	for _, field := range []string{"game", "game.heroHP", "attackSeq", "resourceVersion"} {
		t.Run(field, func(t *testing.T) {
			dungeon := testDungeon(t, "easy")
			switch field {
			case "resourceVersion":
				dungeon.SetResourceVersion("")
			case "game.heroHP":
				unstructured.RemoveNestedField(dungeon.Object, "spec", "game", "heroHP")
			default:
				unstructured.RemoveNestedField(dungeon.Object, "spec", field)
			}
			_, client, api := newTestAPI(t, dungeon)
			response := request(api, http.MethodPost, testPath+"/attacks",
				`{"target":"demo-room-monster-0","seq":0}`, login(t, api))
			if response.Code != http.StatusInternalServerError {
				t.Fatalf("invalid state: %d %s", response.Code, response.Body.String())
			}
			for _, action := range client.Actions() {
				if action.GetVerb() != "get" {
					t.Fatal("invalid state must not be updated")
				}
			}
		})
	}
}

func TestKubernetesErrorsAreExplicitAndSanitized(t *testing.T) {
	for _, tt := range []struct {
		verb string
		path string
		body string
	}{
		{"list", "/api/v1/dungeons", ""},
		{"get", testPath, ""},
		{"create", "/api/v1/dungeons", `{"name":"ok","monsters":1,"difficulty":"easy"}`},
		{"delete", testPath, ""},
		{"update", testPath + "/attacks", `{"target":"demo-room-monster-0","seq":0}`},
	} {
		t.Run(tt.verb, func(t *testing.T) {
			_, client, api := newTestAPI(t, testDungeon(t, "easy"))
			client.PrependReactor(tt.verb, "dungeons", func(action clienttesting.Action) (bool, runtime.Object, error) {
				return true, nil, errors.New("sensitive internal cluster configuration")
			})
			method := map[string]string{"list": "GET", "get": "GET", "create": "POST", "delete": "DELETE", "update": "POST"}[tt.verb]
			response := request(api, method, tt.path, tt.body, login(t, api))
			if response.Code != http.StatusServiceUnavailable || strings.Contains(response.Body.String(), "sensitive") {
				t.Fatalf("Kubernetes error leaked or misclassified: %d %s", response.Code, response.Body.String())
			}
		})
	}
}
