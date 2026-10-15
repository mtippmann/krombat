package handlers

import (
	"net/http"
	"sort"

	"github.com/pnz1990/krombat/backend/internal/k8s"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var inspectableResources = []schema.GroupVersionResource{
	{Group: "game.k8s.example", Version: "v1alpha1", Resource: "heroes"},
	{Group: "game.k8s.example", Version: "v1alpha1", Resource: "monsters"},
	{Group: "game.k8s.example", Version: "v1alpha1", Resource: "bosses"},
	{Group: "game.k8s.example", Version: "v1alpha1", Resource: "treasures"},
	{Group: "game.k8s.example", Version: "v1alpha1", Resource: "loots"},
	{Version: "v1", Resource: "configmaps"},
	{Version: "v1", Resource: "secrets"},
}

const dungeonLabel = "game.k8s.example/dungeon"

func (h *Handler) GetResources(w http.ResponseWriter, r *http.Request) {
	dungeon := h.ownedDungeon(w, r)
	if dungeon == nil {
		return
	}
	candidates := make([]unstructured.Unstructured, 0)
	for _, gvr := range inspectableResources {
		list, err := h.client.Resource(gvr).Namespace(k8s.GameNamespace).
			List(r.Context(), metav1.ListOptions{LabelSelector: dungeonLabel + "=" + dungeon.GetName()})
		if apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			writeK8sError(w, "list dungeon children", err)
			return
		}
		for _, item := range list.Items {
			if item.GetNamespace() != k8s.GameNamespace {
				continue
			}
			if owner := item.GetLabels()[ownerLabel]; owner != "" && owner != demoLogin {
				continue
			}
			if item.GetLabels()[dungeonLabel] != dungeon.GetName() {
				continue
			}
			candidates = append(candidates, item)
		}
	}
	resources := ownedDescendants(dungeon, candidates)
	items := make([]map[string]interface{}, 0, len(resources)+1)
	items = append(items, dungeon.Object)
	for _, resource := range resources {
		items = append(items, resource.Object)
	}
	writeJSON(w, http.StatusOK, items)
}

func ownedDescendants(parent *unstructured.Unstructured, candidates []unstructured.Unstructured) []unstructured.Unstructured {
	known := map[types.UID]*unstructured.Unstructured{}
	if parent.GetUID() != "" {
		known[parent.GetUID()] = parent
	}
	result := make([]unstructured.Unstructured, 0)
	for {
		changed := false
		for i := range candidates {
			child := &candidates[i]
			if child.GetUID() == "" {
				continue
			}
			if _, exists := known[child.GetUID()]; exists {
				continue
			}
			// Upstream kro tracks descendants through ApplySet instance labels,
			// not ownerReferences. Each label identifies the immediate parent.
			labels := child.GetLabels()
			owner, exists := known[types.UID(labels["kro.run/instance-id"])]
			if !exists || labels["kro.run/instance-name"] != owner.GetName() ||
				labels["kro.run/instance-kind"] != owner.GetKind() ||
				labels["kro.run/instance-namespace"] != owner.GetNamespace() ||
				labels["kro.run/instance-group"] != owner.GroupVersionKind().Group ||
				labels["kro.run/instance-version"] != owner.GroupVersionKind().Version {
				continue
			}
			known[child.GetUID()] = child
			result = append(result, *child)
			changed = true
		}
		if !changed {
			break
		}
	}
	sort.Slice(result, func(i, j int) bool {
		a, b := result[i], result[j]
		if a.GetKind() != b.GetKind() {
			return a.GetKind() < b.GetKind()
		}
		return a.GetName() < b.GetName()
	})
	return result
}
