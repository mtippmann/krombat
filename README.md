<div align="center">
  <img src="frontend/public/logo.png" alt="Krombat" width="400">
  <h1>Krombat: vanilla kro on local Kubernetes</h1>
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

### Windows: Docker Desktop

Use **Git Bash**, Docker Desktop in **Linux containers** mode, and enable
Kubernetes in Docker Desktop. Install kubectl, Helm 3, curl, and Python 3
(`python3` or `python` on PATH). The context is `docker-desktop`; the scripts
select it explicitly without changing your current kubeconfig context.

```bash
export KUBECTL_CONTEXT=docker-desktop
bash scripts/local.sh setup
```

Setup installs missing upstream kro v0.9.4 and nginx ingress, but never upgrades
or replaces existing controllers. The pinned ingress-nginx controller is for
this **local-only demo**, not a new public-facing installation.
If port 80 is already occupied, resolve the conflict before continuing.

As Administrator, add this line to
`C:\Windows\System32\drivers\etc\hosts`:

```text
127.0.0.1 krombat.test
```

Then, in the same Git Bash session:

```bash
bash scripts/local.sh build
bash scripts/local.sh deploy
bash scripts/local.sh status
bash tests/run-all.sh
```

**You do not need to push images.** Build uses your active Docker daemon, which
must be Docker Desktop. For its kubeadm cluster, Kubernetes uses the local image
store. For a kind-based Desktop cluster, the script imports the images into each
node container; enable **Show system containers (advanced)** in Docker Desktop
if the script cannot access them. Keep the containerd image store enabled for
kind. No registry, Docker Hub login, or minikube installation is required for
the Docker Desktop path.

Applications retain `imagePullPolicy: Never`; rebuilding and deploying restarts
only the two application deployments. Browser tests use the Docker host gateway,
not Desktop's optional host-networking feature. Windows support must be verified
on your Docker Desktop machine; it cannot be exercised from a Linux-only checkout.

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
For a self-paced workshop on kro concepts and small-company platform design,
see [Docs/workshop](Docs/workshop/README.md).
For authoring these same kro RGDs in CUE and packaging them with Timoni
without Helm, see [Docs/timoni](Docs/timoni/README.md). The companion module
preserves kro graph execution and is validated offline only.
For a 10-minute introduction to CUE, Timoni, Kustomize and Flux, see the
[presentation and speaker notes](Docs/presentation/README.md).

## Layout

| Path | Purpose |
|---|---|
| `backend/` | Go API, local sessions, typed turn logic, upstream CEL playground |
| `frontend/` | React pixel-art game, live resource inspector and teaching panel |
| `manifests/rgds/` | Six upstream-compatible resource graphs |
| `manifests/system/` | One backend, one frontend and nginx ingress |
| `manifests/rbac/` | Namespace-scoped API permissions |
| `timoni/krombat/` | CUE-authored versions of all six RGDs, packaged by Timoni |
| `scripts/local.sh` | Local image build/load, staged deploy, status |
| `tests/` | Local-only guardrails, API/graph assertions and browser journey |

Applications run in `rpg-system`; all game instances and children live in
`rpg-demo`. Same-namespace game resources are easy to inspect and isolate.
The backend can write Dungeons, but cannot write kro's projected child resources.

```bash
bash scripts/local.sh status
kubectl -n rpg-demo get dungeons
kubectl get rgd -l app.kubernetes.io/part-of=krombat
```

## Verify

Build and unit checks run in Docker; integration assertions use the explicitly
selected `minikube` or `docker-desktop` context. Browser gameplay journeys interact only through
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

All cluster tools must explicitly select minikube or docker-desktop.
Deployment and test helpers default to minikube and reject every other context;
tests reject URLs other than `http://krombat.test`.
Existing shared controllers, example graphs and monitoring are not changed.
AWS deployment source and fork-dependent automation have been retired; no cloud
resources were deprovisioned.

The original Git remote is still configured but **must not receive these local
manifests**: its production installation watches that repository. The pre-push
hook permits only the explicitly approved `mtippmann/krombat` fork (`github`
remote) and blocks other destinations.
