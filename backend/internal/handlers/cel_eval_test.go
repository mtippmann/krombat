package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestUpstreamCELBindingsAndLibraries(t *testing.T) {
	evaluator, err := newCELEvaluator()
	if err != nil {
		t.Fatal(err)
	}
	dungeon := testDungeon(t, "easy")
	before := dungeon.DeepCopy()
	for _, tt := range []struct {
		expr string
		want string
	}{
		{`schema.spec.heroClass == "warrior"`, "true"},
		{`schema.spec.game.heroHP`, "200"},
		{`schema.spec.game.bossHP == 0`, "false"},
		{`schema.spec.game.monsterHP.filter(hp, hp > 0).size()`, "2"},
		{`schema.metadata.name`, "demo-room"},
		{`schema.metadata.namespace`, "rpg-demo"},
		{`schema.status.bossPhase`, "phase1"},
		{`cel.bind(hp, schema.spec.game.heroHP, hp * 2)`, "400"},
		{`schema.spec.heroClass.startsWith("war")`, "true"},
		{`quantity("1Gi") == quantity("1024Mi")`, "true"},
		{`random.seededInt(5, 6, schema.metadata.name)`, "5"},
	} {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := evaluator.eval(context.Background(), tt.expr, dungeon)
			if err != nil || result != tt.want {
				t.Fatalf("got %q, %v; want %q", result, err, tt.want)
			}
		})
	}
	if !reflect.DeepEqual(dungeon.Object, before.Object) {
		t.Fatal("CEL playground mutated dungeon")
	}
}

func TestCELRejectsErrorsAndForkOnlyFunctions(t *testing.T) {
	evaluator, err := newCELEvaluator()
	if err != nil {
		t.Fatal(err)
	}
	for _, expr := range []string{
		"", " ", strings.Repeat("x", maxCELExprLength+1),
		"schema.spec.missing.deep + 1", "1 / 0",
		`kstate(schema.spec.game, "heroHP", 0)`,
		strings.Repeat("(", 100) + "1" + strings.Repeat(")", 100),
	} {
		if result, err := evaluator.eval(context.Background(), expr, testDungeon(t, "easy")); err == nil {
			t.Errorf("expected CEL error for %q, got %q", expr, result)
		}
	}
}

func TestCELCostLimit(t *testing.T) {
	evaluator, err := newCELEvaluator()
	if err != nil {
		t.Fatal(err)
	}
	dungeon := testDungeon(t, "easy")
	values := make([]interface{}, 1000)
	for i := range values {
		values[i] = int64(1)
	}
	if err := unstructured.SetNestedSlice(dungeon.Object, values, "spec", "values"); err != nil {
		t.Fatal(err)
	}
	_, err = evaluator.eval(context.Background(),
		`schema.spec.values.all(x, schema.spec.values.all(y, x + y >= 0))`, dungeon)
	if err == nil || !strings.Contains(err.Error(), "cost limit") {
		t.Fatalf("expensive comprehension must hit cost limit: %v", err)
	}
}

func TestCELEndpointContractErrorsAndOwnership(t *testing.T) {
	_, _, api := newTestAPI(t, testDungeon(t, "easy"))
	cookie := login(t, api)
	response := request(api, http.MethodPost, testPath+"/cel-eval",
		`{"expr":"schema.spec.game.heroHP"}`, cookie)
	if response.Code != http.StatusOK {
		t.Fatalf("CEL: %d %s", response.Code, response.Body.String())
	}
	var result celResponse
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil || result.Result != "200" || result.Error != "" {
		t.Fatalf("CEL response: %s, %v", response.Body.String(), err)
	}
	response = request(api, http.MethodPost, testPath+"/cel-eval",
		`{"expr":"schema.spec.missing.deep + 1"}`, cookie)
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil || response.Code != http.StatusOK ||
		result.Result != "" || result.Error == "" {
		t.Fatalf("CEL error response: %d %s, %v", response.Code, response.Body.String(), err)
	}
	dungeon := testDungeon(t, "easy")
	dungeon.SetLabels(map[string]string{ownerLabel: "other"})
	_, _, privateAPI := newTestAPI(t, dungeon)
	response = request(privateAPI, http.MethodPost, testPath+"/cel-eval",
		`{"expr":"schema.spec.game.heroHP"}`, login(t, privateAPI))
	if response.Code != http.StatusNotFound {
		t.Fatalf("unowned dungeon CEL allowed: %d %s", response.Code, response.Body.String())
	}
}

func TestCELRateLimit(t *testing.T) {
	h, _, api := newTestAPI(t, testDungeon(t, "easy"))
	h.celLimit.interval = time.Second
	cookie := login(t, api)
	body := `{"expr":"schema.spec.game.heroHP"}`
	if response := request(api, http.MethodPost, testPath+"/cel-eval", body, cookie); response.Code != http.StatusOK {
		t.Fatalf("initial CEL request: %d", response.Code)
	}
	if response := request(api, http.MethodPost, testPath+"/cel-eval", body, cookie); response.Code != http.StatusTooManyRequests {
		t.Fatalf("CEL rate limit missing: %d", response.Code)
	}
}
