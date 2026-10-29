# 02 - Release graph APIs without Helm

**Goal:** Separate graph authoring, package lifecycle, application lifecycle
and runtime graph reconciliation. Do not apply/delete resources or publish
artifacts during this exercise.

## Package the graph module locally

From the repository root, using an unused archive path:

```bash
timoni mod vet timoni/krombat --name krombat-graphs --namespace rpg-system
timoni mod build ./timoni/krombat --version 0.1.0 \
  --output /tmp/krombat-rgds-0.1.0.oci.tar
```

The artifact packages CUE source and schemas. It is not a game snapshot, a
Helm chart, or a Flux rendered-manifest artifact. It has not been published.

## Design the releases

Read [lifecycle and delivery](../lifecycle.md), then assign ownership:

| Object | Proposed owner |
|---|---|
| Existing upstream kro installation | Shared platform installation, outside this module |
| Six RGDs | Timoni graph instance, or Flux-rendered graph release, not both |
| Generated Dungeon/Hero/etc. CRDs | kro |
| Web/API Deployments, Services, RBAC and ingress | A separate application module/release |
| Dungeon root game state | Go API |
| Hero/Monster/Boss/Treasure/Loot CRs and projections | kro |

Answer:

1. What belongs in graph module values versus application module values
   versus `Dungeon.spec`?
2. Why does another namespace not isolate a second installation of these RGDs?
3. What are the consequences of removing `monster-graph` from a release?
   Why is a manifest rollback not necessarily a data recovery procedure?
4. Why should uninstalling the frontend not remove shared RGDs?
5. What needs to be reviewed before migrating existing objects from the local
   YAML deployment workflow to Timoni inventory?

## Add GitOps only where needed

Draw this delivery path:

```text
Timoni module OCI + release values -> CI renders RGD YAML
  -> Flux manifest OCI artifact -> Flux reconciles RGDs
  -> kro reconciles Dungeon instances and children
```

Compare with direct Timoni `apply`:

1. Which tool stores deployment inventory and performs pruning in each flow?
2. Which tool evaluates the RGD's `forEach`, `includeWhen`, `readyWhen` and CEL?
3. Which artifact is source/configuration versus rendered manifests?
4. Why should neither release pipeline enforce current HP on active Dungeons?
5. Which health and compatibility checks establish that a graph release
   is ready for use as a platform API?

## Deliverable

A release ownership diagram and a short recommendation: use CUE/Timoni for
graph authoring and packaging, state whether continuous Flux delivery is
needed, and identify the application packaging and ownership work not covered
by this offline example.
