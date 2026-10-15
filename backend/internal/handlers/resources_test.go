package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"sort"
	"testing"

	"github.com/pnz1990/krombat/backend/internal/k8s"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clienttesting "k8s.io/client-go/testing"
)

func childObject(parent *unstructured.Unstructured, kind, name, uid string) *unstructured.Unstructured {
	apiVersion := "game.k8s.example/v1alpha1"
	if kind == "ConfigMap" || kind == "Secret" {
		apiVersion = "v1"
	}
	child := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": apiVersion, "kind": kind,
		"metadata": map[string]interface{}{"name": name, "namespace": k8s.GameNamespace, "uid": uid},
	}}
	child.SetLabels(map[string]string{
		dungeonLabel:                 "demo-room",
		"kro.run/instance-id":        string(parent.GetUID()),
		"kro.run/instance-name":      parent.GetName(),
		"kro.run/instance-kind":      parent.GetKind(),
		"kro.run/instance-namespace": parent.GetNamespace(),
		"kro.run/instance-group":     parent.GroupVersionKind().Group,
		"kro.run/instance-version":   parent.GroupVersionKind().Version,
	})
	return child
}

func TestResourcesTraverseUpstreamKroLabelsWithoutOwnerReferences(t *testing.T) {
	parent := testDungeon(t, "easy")
	hero := childObject(parent, "Hero", "demo-room-hero", "hero-uid")
	heroCM := childObject(hero, "ConfigMap", "demo-room-hero-state", "hero-cm-uid")
	loot := childObject(parent, "Loot", "demo-room-loot", "loot-uid")
	lootSecret := childObject(loot, "Secret", "demo-room-loot-data", "loot-secret-uid")
	lootSecret.Object["data"] = map[string]interface{}{"item": "c3dvcmQ="}
	unrelated := childObject(parent, "ConfigMap", "unrelated", "unrelated-uid")
	unrelated.SetLabels(map[string]string{dungeonLabel: "demo-room"})
	forged := childObject(parent, "Secret", "forged-labels", "forged-uid")
	labels := forged.GetLabels()
	labels["kro.run/instance-id"] = "different-parent-uid"
	forged.SetLabels(labels)
	mismatchedKind := childObject(parent, "ConfigMap", "wrong-kind", "wrong-kind-uid")
	labels = mismatchedKind.GetLabels()
	labels["kro.run/instance-kind"] = "Other"
	mismatchedKind.SetLabels(labels)
	ownerRefOnly := childObject(parent, "ConfigMap", "ownerref-only", "ownerref-only-uid")
	ownerRefOnly.SetLabels(map[string]string{dungeonLabel: "demo-room"})
	ownerRefOnly.SetOwnerReferences([]metav1.OwnerReference{{
		APIVersion: parent.GetAPIVersion(), Kind: parent.GetKind(), Name: parent.GetName(), UID: parent.GetUID(),
	}})
	otherOwner := childObject(parent, "Secret", "other-owner", "other-owner-uid")
	labels = otherOwner.GetLabels()
	labels[ownerLabel] = "other"
	otherOwner.SetLabels(labels)
	outside := childObject(parent, "ConfigMap", "outside", "outside-uid")
	outside.SetNamespace("rpg-system")
	otherDungeon := childObject(parent, "ConfigMap", "other-dungeon", "other-dungeon-uid")
	labels = otherDungeon.GetLabels()
	labels[dungeonLabel] = "different-dungeon"
	otherDungeon.SetLabels(labels)
	_, client, api := newTestAPI(t, parent, hero, heroCM, loot, lootSecret,
		unrelated, forged, mismatchedKind, ownerRefOnly, otherOwner, outside, otherDungeon)
	response := request(api, http.MethodGet, testPath+"/resources", "", login(t, api))
	if response.Code != http.StatusOK {
		t.Fatalf("resources: %d %s", response.Code, response.Body.String())
	}
	var items []unstructured.Unstructured
	if err := json.Unmarshal(response.Body.Bytes(), &items); err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 || items[0].GetKind() != "Dungeon" {
		t.Fatalf("resources must include raw parent first: %s", response.Body.String())
	}
	names := []string{}
	for _, item := range items {
		names = append(names, item.GetName())
		if item.GetNamespace() != k8s.GameNamespace || item.Object["metadata"] == nil || item.Object["kind"] == nil {
			t.Fatalf("resource was not a raw scoped Kubernetes object: %v", item.Object)
		}
	}
	sort.Strings(names)
	want := []string{"demo-room", "demo-room-hero", "demo-room-hero-state", "demo-room-loot", "demo-room-loot-data"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("inspector leaked unrelated resources or missed descendants: %v", names)
	}
	allowed := map[string]bool{"dungeons": true}
	for _, resource := range inspectableResources {
		allowed[resource.Resource] = true
	}
	for _, action := range client.Actions() {
		if !allowed[action.GetResource().Resource] || action.GetNamespace() != k8s.GameNamespace {
			t.Fatalf("inspector queried arbitrary resource/scope: %v", action)
		}
		if action.GetVerb() == "list" {
			selector := action.(clienttesting.ListAction).GetListRestrictions().Labels.String()
			if selector != dungeonLabel+"=demo-room" {
				t.Fatalf("child query not dungeon scoped: %s", selector)
			}
		}
	}
}

func TestResourcesRequireParentOwnership(t *testing.T) {
	dungeon := testDungeon(t, "easy")
	dungeon.SetLabels(map[string]string{ownerLabel: "other"})
	_, client, api := newTestAPI(t, dungeon)
	response := request(api, http.MethodGet, testPath+"/resources", "", login(t, api))
	if response.Code != http.StatusNotFound || len(client.Actions()) != 1 {
		t.Fatalf("unowned parent inspector: %d %s, %v", response.Code, response.Body.String(), client.Actions())
	}
}

func TestResourceListFailuresAreNotHidden(t *testing.T) {
	for _, missingKind := range []bool{false, true} {
		t.Run(map[bool]string{false: "access failure", true: "not installed"}[missingKind], func(t *testing.T) {
			_, client, api := newTestAPI(t, testDungeon(t, "easy"))
			client.PrependReactor("list", "heroes", func(action clienttesting.Action) (bool, runtime.Object, error) {
				if missingKind {
					return true, nil, apierrors.NewNotFound(action.GetResource().GroupResource(), "")
				}
				return true, nil, errors.New("permission denied")
			})
			response := request(api, http.MethodGet, testPath+"/resources", "", login(t, api))
			wantStatus := http.StatusServiceUnavailable
			if missingKind {
				wantStatus = http.StatusOK
			}
			if response.Code != wantStatus {
				t.Fatalf("resources: %d %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestResourceAncestryDoesNotInferNamesOrEmptyUIDs(t *testing.T) {
	parent := testDungeon(t, "easy")
	child := childObject(parent, "Hero", "demo-room-hero", "hero-uid")
	parent.SetUID(types.UID(""))
	if result := ownedDescendants(parent, []unstructured.Unstructured{*child}); len(result) != 0 {
		t.Fatalf("empty parent UID matched resources: %v", result)
	}
	parent.SetUID("dungeon-uid")
	child.SetUID("")
	if result := ownedDescendants(parent, []unstructured.Unstructured{*child}); len(result) != 0 {
		t.Fatalf("empty child UID matched resources: %v", result)
	}
}
