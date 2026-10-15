#!/usr/bin/env bash
set -euo pipefail
if [[ "${KUBECTL_CONTEXT:-minikube}" != minikube ]]; then
  echo "Tests may only use minikube." >&2
  exit 1
fi
if [[ "${BASE_URL:-http://krombat.test}" != http://krombat.test ]]; then
  echo "Tests may only use http://krombat.test." >&2
  exit 1
fi
export KUBECTL_CONTEXT=minikube BASE_URL=http://krombat.test
kctl() { kubectl --context minikube "$@"; }
