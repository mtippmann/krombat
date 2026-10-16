#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source scripts/local-common.sh

# These stubs exercise dispatch without contacting Docker or a Kubernetes cluster.
work="$(mktemp -d)"
trap 'rm -f "$work/commands" "$work/output"; rmdir "$work"' EXIT
export COMMAND_LOG="$work/commands"
export STUB_NODE=docker-desktop STUB_SETUP=existing
record() { printf '%s\n' "$*" >> "$COMMAND_LOG"; }
docker() {
  record "docker $*"
  case "$*" in
    'info --format {{.OSType}}') echo linux ;;
    inspect*) echo docker.io/kindest/node:v1.34.1 ;;
    'exec -i '*) cat >/dev/null ;;
  esac
}
kubectl() {
  record "kubectl $*"
  [[ "$1" == --context && "$2" =~ ^(minikube|docker-desktop)$ ]] || return 1
  shift 2
  case "$*" in
    'get nodes -o '*) echo "$STUB_NODE" ;;
    *'get deployment kro -o jsonpath='*) echo registry.k8s.io/kro/kro:v0.9.4 ;;
    *'get deployment kro --ignore-not-found -o name')
      [[ "$STUB_SETUP" != existing ]] || echo deployment.apps/kro ;;
    'get ingressclass nginx --ignore-not-found -o name')
      [[ "$STUB_SETUP" != existing ]] || echo ingressclass.networking.k8s.io/nginx ;;
    'get ingressclass nginx -o jsonpath='*) echo k8s.io/ingress-nginx ;;
    'get namespace kro-system --ignore-not-found -o name')
      [[ "$STUB_SETUP" != collision ]] || echo namespace/kro-system ;;
  esac
  return 0
}
minikube() { record "minikube $*"; }
helm() { record "helm $*"; }
curl() { record "curl $*"; }
export -f record docker kubectl minikube helm curl python_local

absent() {
  if grep -q "$1" "$COMMAND_LOG"; then
    echo "Unexpected command matching: $1" >&2
    exit 1
  fi
}

reject() {
  : > "$COMMAND_LOG"
  if "$@" >"$work/output" 2>&1; then
    echo "Expected rejection: $*" >&2
    exit 1
  fi
  [[ ! -s "$COMMAND_LOG" ]] || {
    echo "Rejected target executed an external command." >&2
    exit 1
  }
}

reject env KUBECTL_CONTEXT=production bash scripts/local.sh build
reject env KUBECTL_CONTEXT=production bash tests/run.sh
reject env BASE_URL=https://example.invalid bash tests/run.sh
reject env KUBECTL_CONTEXT=minikube bash scripts/local.sh setup
reject env KUBECTL_CONTEXT=production bash -c 'python_local -O tests/local-integration.py'
reject env BASE_URL=https://example.invalid bash -c 'python_local -O tests/local-integration.py'

: > "$COMMAND_LOG"
env -u KUBECTL_CONTEXT bash scripts/local.sh build
grep -q '^minikube -p minikube image load krombat-backend:local krombat-frontend:local$' "$COMMAND_LOG"

: > "$COMMAND_LOG"
KUBECTL_CONTEXT=docker-desktop bash scripts/local.sh build
absent '^minikube '
absent '^docker exec '
grep -q '^kubectl --context docker-desktop get nodes ' "$COMMAND_LOG"

: > "$COMMAND_LOG"
STUB_NODE=docker-desktop-control-plane KUBECTL_CONTEXT=docker-desktop bash scripts/local.sh build
grep -q '^docker exec -i docker-desktop-control-plane ctr --namespace k8s.io images import -$' "$COMMAND_LOG"

: > "$COMMAND_LOG"
KUBECTL_CONTEXT=docker-desktop bash scripts/local.sh setup
absent '^helm '

: > "$COMMAND_LOG"
STUB_SETUP=missing KUBECTL_CONTEXT=docker-desktop bash scripts/local.sh setup
grep -q '^helm --kube-context docker-desktop install kro .* --version 0.9.4 ' "$COMMAND_LOG"
grep -q '^helm --kube-context docker-desktop install ingress-nginx .* --version 4.15.1 ' "$COMMAND_LOG"

: > "$COMMAND_LOG"
if STUB_SETUP=collision KUBECTL_CONTEXT=docker-desktop bash scripts/local.sh setup >"$work/output" 2>&1; then
  echo "Setup adopted an existing namespace." >&2
  exit 1
fi
absent '^helm '

KUBECTL_CONTEXT=docker-desktop bash -c '
  source scripts/local-common.sh
  browser_network
  [[ "${BROWSER_NETWORK[*]}" == "--add-host krombat.test:host-gateway" ]]
'
echo "Local script regression checks passed (stubbed; no cluster contacted)."
