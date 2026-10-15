#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../.."
source tests/helpers.sh
docker build -f tests/e2e/Dockerfile -t krombat-browser-tests:local .
output=$(mktemp)
trap 'rm -f "$output"' EXIT
docker run --rm --network host -e BASE_URL=http://krombat.test \
  --add-host "krombat.test:$(minikube -p minikube ip)" \
  krombat-browser-tests:local | tee "$output"
while read -r name; do
  [[ "$name" =~ ^browser-[0-9]+(-hard)?$ ]] || exit 1
  kctl -n rpg-demo wait --for=delete "dungeon/$name" --timeout=90s
  [[ "$(kctl -n rpg-demo get hero,monster,boss,treasure,loot,cm,secret \
    -l "game.k8s.example/dungeon=$name" -o name)" == "" ]]
done < <(sed -n 's/^CLEANUP //p' "$output")
