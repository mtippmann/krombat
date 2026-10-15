#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

if [[ "${KUBECTL_CONTEXT:-minikube}" != minikube ]]; then
  echo "Only KUBECTL_CONTEXT=minikube is allowed." >&2
  exit 1
fi
kctl() { kubectl --context minikube "$@"; }

preflight() {
  kctl get node minikube >/dev/null
  [[ "$(kctl -n kro-system get deployment kro -o jsonpath='{.spec.template.spec.containers[0].image}')" == registry.k8s.io/kro/kro:v0.9.4 ]] || {
    echo "Expected existing upstream kro v0.9.4; will not replace the shared installation." >&2
    exit 1
  }
  kctl get ingressclass nginx >/dev/null
  getent ahostsv4 krombat.test >/dev/null || {
    echo "krombat.test must resolve to minikube before deployment." >&2
    exit 1
  }
  for ns in rpg-demo rpg-system; do
    if [[ -n "$(kctl get namespace "$ns" --ignore-not-found -o name)" ]]; then
      [[ "$(kctl get namespace "$ns" -o jsonpath='{.metadata.labels.app\.kubernetes\.io/part-of}')" == krombat ]] || {
        echo "Refusing to adopt unrelated namespace $ns." >&2; exit 1;
      }
    fi
  done
  for graph in loot hero treasure monster boss dungeon; do
    name="$graph-graph"
    if [[ -n "$(kctl get rgd "$name" --ignore-not-found -o name)" ]]; then
      [[ "$(kctl get rgd "$name" -o jsonpath='{.metadata.labels.app\.kubernetes\.io/part-of}')" == krombat ]] || {
        echo "Refusing to adopt unrelated RGD $name." >&2; exit 1;
      }
    fi
  done
  for crd in dungeons heroes monsters bosses treasures loots; do
    if [[ -n "$(kctl get crd "$crd.game.k8s.example" --ignore-not-found -o name)" ]]; then
      [[ "$(kctl get crd "$crd.game.k8s.example" -o jsonpath='{.metadata.labels.kro\.run/owned}')" == true ]] || {
        echo "Refusing to adopt unrelated CRD $crd.game.k8s.example." >&2; exit 1;
      }
      owner="$(kctl get crd "$crd.game.k8s.example" -o jsonpath='{.metadata.labels.kro\.run/resource-graph-definition-name}')"
      [[ "$(kctl get rgd "$owner" -o jsonpath='{.metadata.labels.app\.kubernetes\.io/part-of}')" == krombat ]] || {
        echo "CRD belongs to unrelated graph $owner." >&2; exit 1;
      }
    fi
  done
}

graph() {
  local name="$1" plural="$2"
  kctl apply --dry-run=server -f "manifests/rgds/$name-graph.yaml" >/dev/null
  kctl apply -f "manifests/rgds/$name-graph.yaml"
  kctl wait --for=condition=Ready "rgd/$name-graph" --timeout=120s
  kctl wait --for=condition=Established "crd/$plural.game.k8s.example" --timeout=60s
}

case "${1:-}" in
  build)
    docker build -t krombat-backend:local backend
    docker build -t krombat-frontend:local frontend
    minikube -p minikube image load krombat-backend:local krombat-frontend:local
    ;;
  deploy)
    preflight
    kctl kustomize manifests >/dev/null
    kctl apply --dry-run=server -f manifests/namespaces.yaml >/dev/null
    kctl apply -f manifests/namespaces.yaml
    kctl apply --dry-run=server -f manifests/rbac/rbac.yaml >/dev/null
    kctl apply -f manifests/rbac/rbac.yaml
    if [[ -z "$(kctl -n rpg-system get secret krombat-session --ignore-not-found -o name)" ]]; then
      # Stream generated secret material directly to Kubernetes, never to a source file.
      openssl rand -hex 32 | tr -d '\n' |
        kctl -n rpg-system create secret generic krombat-session --from-file=SESSION_SECRET=/dev/stdin
    fi
    graph loot loots
    graph hero heroes
    graph treasure treasures
    graph monster monsters
    graph boss bosses
    graph dungeon dungeons
    kctl apply --dry-run=server -f manifests/system >/dev/null
    kctl apply -f manifests/system
    kctl -n rpg-system rollout restart deployment/rpg-backend deployment/rpg-frontend
    kctl -n rpg-system rollout status deployment/rpg-backend --timeout=120s
    kctl -n rpg-system rollout status deployment/rpg-frontend --timeout=120s
    curl --retry 10 --retry-all-errors --retry-delay 1 --connect-timeout 3 --max-time 5 \
      --fail --silent --show-error http://krombat.test/ >/dev/null
    echo "Local demo: http://krombat.test"
    ;;
  status)
    kctl -n rpg-system get deployments,services,ingress
    kctl get rgd -l app.kubernetes.io/part-of=krombat
    kctl -n rpg-demo get dungeons
    ;;
  *)
    echo "Usage: $0 {build|deploy|status}" >&2
    exit 2
    ;;
esac
