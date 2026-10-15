#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source tests/helpers.sh
bash tests/guardrails.sh
docker build --target test -t krombat-backend-tests backend
docker build -t krombat-frontend:local frontend
bash tests/run.sh
bash tests/e2e/run.sh
