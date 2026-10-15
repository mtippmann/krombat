# Krombat local demo

This checkout targets **minikube only**, using unmodified upstream kro v0.9.4.

- Every kubectl command must explicitly pass `--context minikube`. Scripts must
  reject other contexts. Never change the shared current kubeconfig context.
- Never contact the original AWS cluster or public game service, including tests.
- Build and test applications with Docker; run them in minikube, not host dev servers.
- Reuse the existing kro installation and nginx ingress. Do not upgrade or remove
  shared controllers, existing example graphs, monitoring, or unrelated workloads.
- Go owns turn calculations and writes Dungeon `spec.game`. kro owns projected
  Hero, Monster, Boss, Treasure, Loot, ConfigMap and Secret resources.
- `rpg-system` contains the applications; `rpg-demo` contains game resources.
- Keep game parent/child objects in the same namespace for valid ownership.
- Use `http://krombat.test`; no port-forward is needed.
- Local login deliberately provides a shared demo identity, not public authentication.
- Use `scripts/local.sh build`, `scripts/local.sh deploy`, and `tests/run-all.sh`.
- Scope test cleanup to its explicit run label. Never delete all dungeons.
- Do not publish to the original remote or merge automatically: the original
  production installation watches that repository.
- Preserve untracked guide files and other user changes.
