package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/cel-go/cel"
	krocel "github.com/kubernetes-sigs/kro/pkg/cel"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const maxCELExprLength = 2000

type celEvaluator struct {
	env *cel.Env
}

func newCELEvaluator() (*celEvaluator, error) {
	options := append([]cel.EnvOption{}, krocel.BaseDeclarations()...)
	options = append(options,
		cel.Variable("schema", cel.MapType(cel.StringType, cel.DynType)),
		cel.ParserRecursionLimit(64),
	)
	env, err := cel.NewEnv(options...)
	if err != nil {
		return nil, err
	}
	return &celEvaluator{env: env}, nil
}

func (e *celEvaluator) eval(ctx context.Context, expr string, dungeon *unstructured.Unstructured) (string, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "", fmt.Errorf("expression is required")
	}
	if len(expr) > maxCELExprLength {
		return "", fmt.Errorf("expression too long (max %d bytes)", maxCELExprLength)
	}
	ast, issues := e.env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return "", issues.Err()
	}
	program, err := e.env.Program(ast, cel.CostLimit(10000), cel.InterruptCheckFrequency(100))
	if err != nil {
		return "", err
	}
	bindings := make(map[string]interface{}, 3)
	for _, field := range []string{"spec", "status", "metadata"} {
		value, found, err := unstructured.NestedMap(dungeon.Object, field)
		if err != nil {
			return "", fmt.Errorf("invalid dungeon %s", field)
		}
		if !found {
			value = map[string]interface{}{}
		}
		bindings[field] = value
	}
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	out, _, err := program.ContextEval(ctx, map[string]interface{}{"schema": bindings})
	if err != nil {
		return "", err
	}
	return fmt.Sprint(out), nil
}

type celResponse struct {
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}

func (h *Handler) CELEval(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expr string `json:"expr"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	dungeon := h.ownedDungeon(w, r)
	if dungeon == nil {
		return
	}
	result, err := h.cel.eval(r.Context(), req.Expr, dungeon)
	response := celResponse{Result: result}
	if err != nil {
		response.Error = err.Error()
	}
	writeJSON(w, http.StatusOK, response)
}
