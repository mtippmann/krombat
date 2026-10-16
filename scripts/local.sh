#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

source scripts/local-common.sh

check_kro() {
  [[ "$(kctl -n kro-system get deployment kro -o jsonpath='{.spec.template.spec.containers[0].image}')" == registry.k8s.io/kro/kro:v0.9.4 ]] || {
    echo "Expected upstream kro v0.9.4; will not replace an existing installation." >&2
    exit 1
  }
}

setup() {
  [[ "$KUBECTL_CONTEXT" == docker-desktop ]] || {
    echo "Setup is only allowed for docker-desktop; minikube prerequisites are shared." >&2
    exit 1
  }
  command -v helm >/dev/null || { echo "Install Helm 3 before running setup." >&2; exit 1; }
  kctl wait --for=condition=Ready nodes --all --timeout=120s
  # Check all collisions before installing either controller.
  local install_kro=false install_nginx=false
  if [[ -n "$(kctl -n kro-system get deployment kro --ignore-not-found -o name)" ]]; then
    check_kro
  elif [[ -n "$(kctl get namespace kro-system --ignore-not-found -o name)" ]]; then
    echo "Existing kro-system namespace without deployment/kro; refusing to adopt it." >&2
    exit 1
  else
    install_kro=true
  fi
  if [[ -n "$(kctl get ingressclass nginx --ignore-not-found -o name)" ]]; then
    [[ "$(kctl get ingressclass nginx -o jsonpath='{.spec.controller}')" == k8s.io/ingress-nginx ]] || {
      echo "Existing nginx class belongs to another controller; refusing to replace it." >&2
      exit 1
    }
    echo "Reusing the existing nginx ingress class without changing its controller."
  elif [[ -n "$(kctl get namespace ingress-nginx --ignore-not-found -o name)" ]]; then
    echo "Existing ingress-nginx namespace without ingressclass/nginx; refusing to adopt it." >&2
    exit 1
  else
    install_nginx=true
  fi
  if "$install_kro"; then
    hctl install kro oci://registry.k8s.io/kro/charts/kro --version 0.9.4 \
      --namespace kro-system --create-namespace --wait --timeout 180s
  fi
  if "$install_nginx"; then
    echo "Installing ingress-nginx for this local-only demo; do not expose it publicly."
    hctl install ingress-nginx ingress-nginx \
      --repo https://kubernetes.github.io/ingress-nginx --version 4.15.1 \
      --namespace ingress-nginx --create-namespace \
      --set controller.service.type=LoadBalancer --wait --timeout 180s
  fi
  check_kro
  curl --noproxy '*' --connect-timeout 3 --max-time 5 --silent --show-error \
    --resolve krombat.test:80:127.0.0.1 http://krombat.test/ >/dev/null || {
      echo "Cannot reach ingress on localhost:80. Check the controller service and port conflicts." >&2
      exit 1
    }
  echo 'Add this entry to C:\Windows\System32\drivers\etc\hosts as Administrator:'
  echo "127.0.0.1 krombat.test"
}

load_images() {
  if [[ "$KUBECTL_CONTEXT" == minikube ]]; then
    minikube -p minikube image load krombat-backend:local krombat-frontend:local
    return
  fi
  local nodes node image
  nodes="$(kctl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}')"
  [[ -n "$nodes" ]] || { echo "Docker Desktop Kubernetes has no nodes." >&2; exit 1; }
  while IFS= read -r node; do
    node="${node%$'\r'}"
    # kubeadm shares Docker Desktop's local image store; kind has per-node stores.
    if [[ "$node" != docker-desktop ]]; then
      image="$(dkr inspect --type container --format '{{.Config.Image}}' "$node")" || {
        echo "Cannot access kind node $node. Enable Docker Desktop's Show system containers setting." >&2
        exit 1
      }
      [[ "$image" == *kindest/node:* ]] || {
        echo "Refusing to import images into unrecognized node container $node ($image)." >&2
        exit 1
      }
      dkr image save krombat-backend:local krombat-frontend:local |
        dkr exec -i "$node" ctr --namespace k8s.io images import -
    fi
  done <<< "$nodes"
}

preflight() {
  kctl wait --for=condition=Ready nodes --all --timeout=120s
  check_kro
  kctl get ingressclass nginx >/dev/null
  check_local_address
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
  setup)
    setup
    ;;
  build)
    if [[ "$KUBECTL_CONTEXT" == docker-desktop ]]; then
      [[ "$(dkr info --format '{{.OSType}}')" == linux ]] || {
        echo "Docker Desktop must be running Linux containers." >&2; exit 1;
      }
    fi
    dkr build -t krombat-backend:local backend
    dkr build -t krombat-frontend:local frontend
    load_images
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
      python_local - <<'PY' | kctl create -f -
import json
import secrets

print(json.dumps({
    "apiVersion": "v1", "kind": "Secret",
    "metadata": {"name": "krombat-session", "namespace": "rpg-system"},
    "type": "Opaque", "stringData": {"SESSION_SECRET": secrets.token_hex(32)}
}))
PY
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
    curl --noproxy '*' --retry 10 --retry-all-errors --retry-delay 1 --connect-timeout 3 --max-time 5 \
      --fail --silent --show-error http://krombat.test/ >/dev/null
    echo "Local demo: http://krombat.test"
    ;;
  status)
    kctl -n rpg-system get deployments,services,ingress
    kctl get rgd -l app.kubernetes.io/part-of=krombat
    kctl -n rpg-demo get dungeons
    ;;
  *)
    echo "Usage: KUBECTL_CONTEXT=minikube|docker-desktop $0 {setup|build|deploy|status}" >&2
    exit 2
    ;;
esac
