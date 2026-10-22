# Helm-free lifecycle: Timoni delivers, kro reconciles

The company adoption question is not "CUE instead of kro." It is "CUE for
readable, reusable graph authoring; Timoni for versioned releases; kro for
the platform API and its runtime graph."

## Keep three lifecycles separate

| Release/resource | Owner and responsibility | Do not conflate with |
|---|---|---|
| kro controller and upstream CRDs | Existing shared platform installation; this example reuses upstream v0.9.4 | The Krombat graph package must not upgrade/install kro. |
| Platform graph module | Timoni manages six RGD objects. kro generates their APIs and reconciles instances. | A namespace-scoped application release or one game session |
| Web/API application module | A separate Timoni module can manage Deployments, Services, app RBAC and ingress in `rpg-system` | Shared RGDs/controllers or the game objects kro creates |
| Dungeon instances | Go API writes the root `Dungeon.spec.game`; kro composes and projects children in `rpg-demo` | A release manager continually resetting HP from static values |

The implemented module is the **graph module**. The application module in
this design is a future packaging step, not implemented here. Its starting
manifests are `manifests/system/` and `manifests/rbac/`; preserve the current
image names, local image loading, namespace boundaries, ingress and local
session-secret handling when packaging them. Do not have two release
managers own those same objects.

## The Helm values equivalent

Timoni's `values` interface is analogous to Helm chart values, with CUE
constraints/defaults rather than Go-template string substitution. Users can
supply CUE, YAML or JSON values. Module, Bundle and Instance roughly
correspond to a chart, a grouped release definition, and an installed release;
their exact lifecycle semantics are not identical to Helm's.

For this graph module:

```cue
values: {
	graphLabels: "platform.example/team": "workshop"
	dungeonDefaults: {
		monsters:   5
		difficulty: "normal"
	}
}
```

These are **platform API release inputs**. An application module would have a
different interface, for example image references, resource requests and
replicas. A player chooses difficulty through the Dungeon API. Game HP and
kro-generated status belong to runtime objects, not release values.

## Direct Timoni lifecycle, without a chart

Timoni can apply Kubernetes objects from a local or OCI module directly; it
does not need Helm to install Deployments, Services, RGDs or other resources.
The RGD module itself includes no HelmRelease and no controller installation.

| Stage | Timoni operation | What it means for an RGD release |
|---|---|---|
| Author/check | `fmt`, `mod vet`, `build` | Validate values and the outer RGD schema, inspect generated graph changes, run parity/behavior checks. |
| Package | `mod build` | Write a local versioned OCI module archive. No deployment or registry is involved. |
| Publish/promote | `mod push`, signing/verification, version/digest pinning | Distribute reviewed graph source and schemas. Registry and trust configuration need separate approval. |
| Preview | `apply --dry-run --diff` | Query Kubernetes and preview changes to RGD objects; this is not an offline command. |
| Install/upgrade | `apply` | Server-side apply the graphs, track the instance inventory and wait for health. kro compiles/activates the graph changes. |
| Inspect | `list`, `inspect module/values/resources`, `status` | Inspect the graph package. Use kro's CRs/conditions and game inspector for runtime graph behavior. |
| Revert release | Apply a previous module digest and its values | Revert release inputs if the API/schema change is reversible; this does not restore game data or reverse destructive API migrations. |
| Retire | `delete`, or prune objects removed from the next revision | Removes managed RGDs, which can affect generated APIs and all their instances; not a routine app reset. |

Timoni stores instance inventory in a `timoni.<instance>` Secret and prunes
previously managed objects absent from a new revision. Its apply wait/health
checks differ from kro's `readyWhen` inside the RGD. For a future live pilot,
verify RGD condition/observed-generation behavior and configure suitable
health checks before relying on upgrade gates. Offline vet/build does not
prove a graph compiled successfully in kro.

The executable path for **this** checkout is intentionally limited to:

```bash
timoni mod vet timoni/krombat --name krombat-graphs --namespace rpg-system
timoni build krombat-graphs ./timoni/krombat --namespace rpg-system
timoni mod build ./timoni/krombat --version 0.1.0 \
  --output /tmp/krombat-rgds-0.1.0.oci.tar
```

Choose an unused archive path. No lifecycle-changing commands are run here.
Any future Kubernetes operation must explicitly select the validated
`minikube` or `docker-desktop` context, never the shared current context.

## Ownership and upgrade hazards

**RGDs are cluster-scoped.** Their fixed names and generated API identities
are not isolated by a different Timoni instance name or namespace. Applying
this module over the existing `*-graph` objects is an ownership migration,
not a harmless side-by-side install. Timoni can adopt existing resources;
do not allow that implicitly in the shared demo.

For a company pilot, choose one owner for each graph package, review
compatibility of the generated CRD schema and all existing instances, and
plan resource adoption explicitly. Keep graph names, API identities and
resource IDs stable. Removing a graph from a release can trigger pruning and
impact more than one application. A rollback of manifests cannot restore
data lost through API/resource deletion.

Keep the graph/API module and application module separately versioned so
restarting or uninstalling the web frontend cannot remove a shared platform
API. If a Bundle composes them, account explicitly for dependency/readiness
ordering and the effects of bundle-wide removal. Do not create duplicate
ownership of RGDs, generated CRDs, or kro-managed children.

## Add Flux second

Two distinct Flux uses are worth separating:

1. **GitOps delivery of these graphs:** CI builds the Timoni module with
   release values, publishes the rendered RGD YAML as a Flux OCI artifact,
   and Flux reconciles that artifact. Flux owns the RGDs; kro still owns
   game-instance reconciliation. Do not also `timoni apply` the same RGDs.
2. **Timoni orchestrates Flux APIs:** other Timoni modules can configure
   Flux controllers or HelmRelease resources. That is a separate integration
   and is unnecessary for a Helm-free RGD/application module.

The first flow is:

```text
CUE RGD source + release values
  -> Timoni build in CI
  -> rendered RGD manifest artifact in OCI
  -> Flux OCIRepository/Kustomization
  -> ResourceGraphDefinitions
  -> kro-generated APIs and reconciliation of game instances
```

The **Timoni module OCI artifact** contains source and schemas for reuse.
The **Flux manifest OCI artifact** contains rendered YAML. They are not the
same delivery artifact. In the documented Flux-templating workflow, Flux
owns apply/prune/wait rather than Timoni instance inventory; Timoni's
direct-CLI Bundle lifecycle, runtime injection and multi-cluster apply are
not provided by an in-cluster Timoni controller.

No Flux installation, registry push, Helm migration or cluster apply is in
scope here.

## Sources

- [Timoni concepts and instance inventory](https://timoni.sh/concepts)
- [Modules](https://timoni.sh/module) and [custom resources](https://timoni.sh/cue/module/custom-resources)
- [Apply behavior and adoption](https://timoni.sh/cue/module/apply-behavior)
- [Custom health checks](https://timoni.sh/cue/module/health-checks)
- [OCI publishing](https://timoni.sh/cue/module/publishing) and [signing](https://timoni.sh/cue/module/signing)
- [Bundles](https://timoni.sh/bundle) and [runtime values](https://timoni.sh/bundle-runtime)
- [Flux GitOps](https://timoni.sh/gitops-flux) and [Helm interoperability](https://timoni.sh/flux-helm-interop)
