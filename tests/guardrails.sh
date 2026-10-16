#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source tests/helpers.sh
if grep -REn 'learn-kro\.eks\.aws\.dev|arn:aws:|dkr\.ecr\.|aws eks|aws-actions/|pnz1990/kro |kstate\(' \
  manifests backend frontend/src frontend/nginx.conf scripts tests .github/workflows \
  --exclude=guardrails.sh --exclude=go.sum --exclude='*_test.go' \
  --exclude-dir=.cache --exclude-dir=node_modules --exclude-dir=dist; then
  echo "Retired production/fork reference found in active code." >&2
  exit 1
fi
if grep -REn '^      state:|specPatch:|stateWrite:' manifests/rgds; then
  echo "Fork state nodes are not allowed." >&2
  exit 1
fi
if grep -RIn 'kubectl ' scripts tests --exclude=guardrails.sh --exclude=local-integration.py --exclude=local-script.sh |
  grep -vF -- 'kubectl --context "$KUBECTL_CONTEXT"'; then
  echo "kubectl calls must explicitly select the validated local context." >&2
  exit 1
fi
if grep -q '^replace .*kro' backend/go.mod; then
  echo "Backend must use upstream kro." >&2
  exit 1
fi
for script in scripts/*.sh tests/*.sh tests/e2e/run.sh .githooks/pre-push; do
  bash -n "$script"
done
echo "Local-only architecture guardrails passed."
