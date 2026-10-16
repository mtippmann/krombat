#!/usr/bin/env bash

case "${KUBECTL_CONTEXT:-minikube}" in
  minikube|docker-desktop) export KUBECTL_CONTEXT="${KUBECTL_CONTEXT:-minikube}" ;;
  *) echo "Only minikube and docker-desktop contexts are allowed." >&2; exit 1 ;;
esac

kctl() { MSYS_NO_PATHCONV=1 kubectl --context "$KUBECTL_CONTEXT" "$@"; }
dkr() { MSYS_NO_PATHCONV=1 docker "$@"; }
hctl() { MSYS_NO_PATHCONV=1 helm --kube-context "$KUBECTL_CONTEXT" "$@"; }

python_local() {
  local command
  for command in python3 python; do
    if "$command" --version >/dev/null 2>&1; then
      "$command" "$@"
      return
    fi
  done
  echo "Python 3 is required (python3 or python on PATH)." >&2
  return 1
}

check_local_address() {
  local address=127.0.0.1
  if [[ "$KUBECTL_CONTEXT" == minikube ]]; then
    address="$(minikube -p minikube ip)"
  fi
  python_local - "$address" <<'PY'
import socket
import sys

try:
    addresses = {item[4][0] for item in socket.getaddrinfo("krombat.test", 80, socket.AF_INET)}
except socket.gaierror as error:
    sys.exit(f"krombat.test lookup failed: {error}. Add a hosts-file entry for {sys.argv[1]}.")
if addresses != {sys.argv[1]}:
    sys.exit(f"krombat.test must resolve only to {sys.argv[1]}, found {sorted(addresses)}.")
PY
}

browser_network() {
  if [[ "$KUBECTL_CONTEXT" == minikube ]]; then
    BROWSER_NETWORK=(--network host --add-host "krombat.test:$(minikube -p minikube ip)")
  else
    BROWSER_NETWORK=(--add-host krombat.test:host-gateway)
  fi
}
