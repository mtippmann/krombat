# kro: APIs, resource composition and reconciliation

A self-paced workshop using the local Krombat demo to explore kro as a
Kubernetes platform-engineering tool. The game is a concrete resource graph to
inspect, not an application to extend. No kro fork, Go changes, or custom
controller implementation is required.

## Learning outcomes

By completing the modules, you should be able to:

- Explain how a ResourceGraphDefinition (RGD) exposes a typed Kubernetes API
  and composes resources from an instance.
- Read the schema, resource templates, CEL expressions, readiness conditions,
  and status projections in an RGD.
- Recognize composition, collection expansion, conditional resources,
  derived status, and reconciliation in a working example.
- Compare resource-composition patterns, their operational requirements
  and the abstractions they introduce.
- Describe what the game does **not** demonstrate, without mistaking a demo
  architecture for a production recommendation.

## Before you begin

You need basic Kubernetes familiarity (CRs, namespaces, ConfigMaps, Secrets,
and controllers), a browser, and access to the local Krombat environment.
Follow the repository [start instructions](../../README.md#start) to build and
deploy the demo. On Docker Desktop, follow the Windows/Docker Desktop setup and
set `KUBECTL_CONTEXT=docker-desktop` for repository scripts. The default local
context is minikube. Never switch the shared kubeconfig context or point these
exercises at a cloud cluster.

Open `http://krombat.test`, select **Enter local demo**, and create a dungeon
with three monsters. Use a unique name such as `team-workshop`. Complete the
activities in order, or use them as standalone labs. The game UI's resource
inspector and **What is kro doing?** panel are sufficient; optional commands
below use an explicitly named local context.

```bash
bash scripts/local.sh status
kubectl --context minikube get rgd -l app.kubernetes.io/part-of=krombat
kubectl --context minikube -n rpg-demo get dungeon team-workshop -o yaml
```

For Docker Desktop, replace `--context minikube` with
`--context docker-desktop` in the kubectl examples. Do not edit resources
during the workshop unless an exercise explicitly asks you to. In particular,
do not reapply `Docs/demo/dungeon-demo.yaml` over an active game: it resets
desired game state.

## Modules

| Module | Focus | Evidence / output |
|---|---|---|
| [01 - Observe the graph](exercises/01-observe-the-graph.md) | API, composition, reconciliation and lifecycle | Annotated map of one Dungeon and its composed resources |
| [02 - Read RGDs and CEL](exercises/02-read-rgds-and-cel.md) | Schema, templates, `forEach`, `includeWhen`, readiness, status and expressions | Answers grounded in the current RGD YAML and CEL playground |
| [03 - Platform design review](exercises/03-platform-design-review.md) | API boundaries, operations and alternatives | Technical comparison with assumptions, capabilities and costs |

Self-checks are in [reference answers](solutions/answers.md). Try each module
before opening them.

## What this game teaches - and what it does not

The current demo uses unmodified upstream kro v0.9.4. Its six RGDs define
Dungeon, Hero, Monster, Boss, Treasure and Loot APIs. A Dungeon graph composes
the game resources; child graphs compose ConfigMaps and demo Secrets. The
examples show generated APIs, typed schema constraints, `forEach`,
`includeWhen`, `readyWhen`, CEL-derived status, and reconciliation of
desired resources.

The boundary is intentional: Go validates and computes turns, then writes
`Dungeon.spec.game`. kro projects that desired state into child resources and
evaluates CEL for derived state. The CEL playground evaluates read-only
expressions against the latest Dungeon data; it does not update the game or
run inside the RGD controller. The original fork's state-write combat engine
is not present. This workshop does not ask you to implement either Go behavior
or kro itself.

Use the game to discuss platform abstractions, not to claim that all business
logic belongs in Kubernetes. The demo does not establish production
authentication, tenant isolation, high-throughput state storage, or a complete
platform operating model. Its shared local identity, HTTP endpoint, and trophy
Secrets are demo-only.

## Cleanup

Delete only the dungeon you created, using the game's **Delete dungeon**
action. If you chose a different name, substitute it for `team-workshop` below.
Verify by name if desired:

```bash
kubectl --context minikube -n rpg-demo get dungeon team-workshop
```

Do not delete all dungeons, namespaces, RGDs, or shared controllers. kro owns
cleanup of the resources composed for a deleted Dungeon.
