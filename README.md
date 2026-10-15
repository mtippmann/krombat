<div align="center">
  <img src="frontend/public/logo.png" alt="Krombat" width="400">
  <h1>Krombat: vanilla kro on minikube</h1>
  <p>A pixel-art dungeon and a real Kubernetes resource graph.</p>
</div>

This checkout is a local presentation demo, converted from the AWS/fork-based
original. It uses **unmodified upstream kro v0.9.4**, not a custom controller.
No AWS account, GitHub OAuth app, database, Argo CD, or cloud registry is needed.

**The boundary matters:** Go computes turns and writes `Dungeon.spec.game`.
kro projects that desired state into Hero, Monster, Boss, Treasure and Loot
resources, derives entity states and boss phases with CEL, and reconciles
ConfigMaps and Secrets. Vanilla kro does not provide the original fork's
state-write combat engine.

## Start

Prerequisites: Docker, minikube, kubectl, curl, OpenSSL, Python 3, and a running
minikube profile named `minikube` with nginx ingress and upstream kro v0.9.4.
The existing shared kro installation is reused, never replaced. Public image
and dependency downloads require internet during builds.

`krombat.test` must resolve to `minikube -p minikube ip` on your host.
On this machine it already does. The app is reached through ingress, not
port-forward or a host development server.

```bash
git config --local core.hooksPath .githooks
bash scripts/local.sh build
bash scripts/local.sh deploy
```

Open **http://krombat.test** and select **Enter local demo**. Create a warrior
dungeon, defeat the monsters to unlock the boss, observe its phases, defeat it,
and open the chest. Delete the dungeon to reset or create another named run.

All players share the explicit `demo` identity. HTTP cookies are signed with a
locally generated Kubernetes Secret. **This is not public or multi-user
authentication; do not expose it on a public ingress.**

## What to present

| Feature | What actually happens |
|---|---|
| Generated API | Six RGDs register typed Dungeon/Hero/Monster/Boss/Treasure/Loot CRDs |
| Composition | A Dungeon creates child CRs reconciled by their respective graphs |
| Dynamic fan-out | `forEach` creates one Monster per indexed HP entry |
| CEL projections | HP drives alive/dead state and boss phases at 50% and 25% |
| Conditional resources | `includeWhen` creates Loot after kills and the Treasure Secret after opening |
| Readiness and aggregation | A real `readyWhen` checks the hero projection; Dungeon status summarizes children |
| Deterministic CEL randomness | Loot Secrets contain a name-seeded trophy token |
| Reconciliation | Edit/delete a demo-owned ConfigMap; kro restores desired state |
| Lifecycle | ApplySet inventory and kro finalizers remove children when a Dungeon is deleted |

Upstream v0.9.4 uses **ApplySet labels and inventory**, not automatic ownerReferences,
for these managed children. Inspect `kro.run/instance-id`,
`applyset.kubernetes.io/part-of`, and `kro.run/finalizer`. Do not interpret missing
ownerReferences as unmanaged resources or claim Kubernetes garbage collection is
performing kro's inventory cleanup.

See [the presentation walkthrough](Docs/demo/DEMO.md) for Radar/k9s commands and
a repeatable scenario, and [the runbook](Docs/runbook.md) for troubleshooting.

## Layout

| Path | Purpose |
|---|---|
| `backend/` | Go API, local sessions, typed turn logic, upstream CEL playground |
| `frontend/` | React pixel-art game, live resource inspector and teaching panel |
| `manifests/rgds/` | Six upstream-compatible resource graphs |
| `manifests/system/` | One backend, one frontend and nginx ingress |
| `manifests/rbac/` | Namespace-scoped API permissions |
| `scripts/local.sh` | Local image build/load, staged deploy, status |
| `tests/` | Local-only guardrails, API/graph assertions and browser journey |

Applications run in `rpg-system`; all game instances and children live in
`rpg-demo`. Same-namespace game resources are easy to inspect and isolate.
The backend can write Dungeons, but cannot write kro's projected child resources.

```bash
bash scripts/local.sh status
kubectl --context minikube -n rpg-demo get dungeons
kubectl --context minikube get rgd -l app.kubernetes.io/part-of=krombat
```

## Verify

Build and unit checks run in Docker; integration assertions use explicit
`kubectl --context minikube`. Browser gameplay journeys interact only through
the UI. Integration cleanup targets its exact test run, never all dungeons.

```bash
bash tests/run-all.sh
# Or just the deployed API/graph and browser journeys:
bash tests/run.sh
bash tests/e2e/run.sh
```

The supported game is one room and one warrior with three difficulties.
Loot is a visible trophy, not equippable gear. Abilities, modifiers, status
effects, second rooms, progression and social features have been removed.

## Safety

All cluster tools must explicitly select minikube. Deployment and test helpers
reject other contexts; tests reject URLs other than `http://krombat.test`.
Existing shared controllers, example graphs and monitoring are not changed.
AWS deployment source and fork-dependent automation have been retired; no cloud
resources were deprovisioned.

The original Git remote is still configured but **must not receive these local
manifests**: its production installation watches that repository. The pre-push
hook blocks publication. Choose a separate destination explicitly if publishing.
