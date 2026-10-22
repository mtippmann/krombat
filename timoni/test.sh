#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

export KUBECONFIG=/dev/null

timoni fmt --diff timoni/krombat
timoni mod vet timoni/krombat --name krombat-graphs --namespace rpg-system
timoni mod vet timoni/krombat --debug --name krombat-graphs --namespace rpg-system
python3 timoni/tests/test_graphs.py
