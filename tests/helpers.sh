#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/../scripts/local-common.sh"
if [[ "${BASE_URL:-http://krombat.test}" != http://krombat.test ]]; then
  echo "Tests may only use http://krombat.test." >&2
  exit 1
fi
export BASE_URL=http://krombat.test
