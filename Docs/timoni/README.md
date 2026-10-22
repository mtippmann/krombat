# Easier kro RGD authoring with CUE and Timoni

**The goal is to maintain kro graphs more easily, not replace kro.**
Krombat stays a kro game. The CUE module authors the same six RGDs, and
Timoni provides their package/release lifecycle without Helm.

```text
Platform author
  CUE graph definitions + shared helpers + typed release values
    -> Timoni build/vet -> actual ResourceGraphDefinition YAML
    -> versioned Timoni OCI module -> delivery to Kubernetes
                                       |
                                       v
Application user/API -> Dungeon CR -> kro -> Hero/Monster/Boss/Treasure/Loot
                                       -> ConfigMaps/Secrets + live status
```

The generated RGDs still contain `${...}` CEL, `forEach`, `includeWhen`,
`readyWhen`, SimpleSchema and status projections. **kro still executes all
of those.** Go still computes turns and writes `Dungeon.spec.game`.

## Start with the actual example

Open these side by side:

| Existing YAML | CUE authoring counterpart |
|---|---|
| [`monster-graph.yaml`](../../manifests/rgds/monster-graph.yaml) | [`templates/monster.cue`](../../timoni/krombat/templates/monster.cue) |
| [`boss-graph.yaml`](../../manifests/rgds/boss-graph.yaml) | [`templates/boss.cue`](../../timoni/krombat/templates/boss.cue) |
| [`dungeon-graph.yaml`](../../manifests/rgds/dungeon-graph.yaml) | [`templates/dungeon.cue`](../../timoni/krombat/templates/dungeon.cue) |

The [YAML-to-CUE walkthrough](rgd-mapping.md) shows the Monster graph,
its shared helpers and its generated YAML. It explains exactly what is
evaluated by CUE now and what remains for kro later.

Run from the repository root with Timoni 0.34.0:

```bash
timoni build krombat-graphs ./timoni/krombat --namespace rpg-system
bash timoni/test.sh
```

Build emits six RGDs, not a precomputed game resource snapshot. Checks need
Python 3/PyYAML 6.x as well as Timoni, but no cluster or standalone CUE CLI.
They compare the generated graphs against all six existing YAMLs, including
every CEL expression and resource template. The one normalization is
`spec.schema.scope: Namespaced`, explicitly written rather than relying on
the API default.

The existing YAML/deployment scripts and running demo are not changed.
Do not apply this prototype over the existing graphs or publish it to the
original remote.

## What improves for a platform team

| Maintenance task | CUE/Timoni approach in this example |
|---|---|
| Keep graph metadata and API identity consistent | One `#RGD` envelope uses the upstream RGD schema and fixes group, version, namespaced instance scope and required label. |
| Avoid repeating state ConfigMaps | `#StateResource` supplies the template envelope; Hero, Monster, Boss, Treasure and Loot provide their graph-specific data. |
| Avoid copying defeated-entity Loot logic | `#LootResource` shares the runtime `includeWhen`, naming and child-CR template between Monster and Boss; rarity stays graph-specific. |
| Keep table views consistent | Reuse `#HPColumn`, `#StateColumn` and `#ReadyColumn`. |
| Offer controlled release configuration | Typed module values constrain API defaults and release labels; incompatible values fail CUE unification. |
| Review changes safely | Render YAML and compare with the original graphs before packaging; keep CEL, resource IDs and API compatibility visible in review. |
| Distribute and promote a release | Build a versioned OCI module; promote a pinned version/digest with reviewed values. No Helm chart or Go-template layer is required. |

There is a tradeoff: helpers reduce repeated structure but introduce
indirection. Keep graph-specific CEL and dependency IDs visible in the graph
file. Do not hide the whole graph behind a generic abstraction just to reduce
line count. Changing a shared helper can affect multiple RGDs, so review the
rendered output and run parity/behavior checks.

## Values, lifecycle and Flux

The [lifecycle guide](lifecycle.md) separates three owners:

- **Platform API release:** Timoni or Flux delivers RGDs; kro owns the
  generated APIs and game projections.
- **Application release:** a separate Timoni module can manage web/API
  Deployments, Services, RBAC and ingress without Helm.
- **Game-instance lifecycle:** the Go API creates/updates/deletes Dungeons;
  kro reconciles their children. Neither Timoni nor Flux should enforce
  gameplay HP fields on active Dungeons.

The prototype implements the first module. It does **not** install a Timoni
application release, Flux or kro, and does not migrate existing ownership.
The application-release design and Flux pipeline are explanations, not
deployed or live-tested features.

## VS Code and author experience

Use the [official CUE extension](https://marketplace.visualstudio.com/items?itemName=cuelangorg.vscode-cue)
(`cuelangorg.vscode-cue`). It provides syntax highlighting and CUE language
server integration. Open `timoni/krombat` as a workspace folder so its
`cue.mod` is the module root; keep the original YAML open beside the graph.
The [extension documentation](https://github.com/cue-lang/vscode-cue)
describes `cue.cueCommand` for selecting the CUE executable.

The extension needs a separate CUE executable for LSP. Prefer a CUE version
matching the Timoni engine (`timoni version`; this module uses CUE 0.17.1).
Timoni embeds its own engine/formatter, so its build/vet/fmt commands do not
need that executable.

For feedback while authoring:

1. Edit `templates/monster.cue` or a narrowly scoped shared helper.
2. Use CUE language-server diagnostics for syntax, references and
   constraints. The imported RGD schema provides type information for the
   RGD structure, not proof that embedded CEL or game behavior is correct.
3. Run `timoni fmt --diff timoni/krombat` and `timoni mod vet timoni/krombat`.
4. Render the six RGDs and inspect the resulting YAML diff; run
   `bash timoni/test.sh` to catch accidental behavior changes.

There is no dedicated Timoni extension documented on its site. Generic
community CUE/YAML tooling can complement this workflow but is not a
kro-aware graph debugger. Timoni's [agent skills and docs MCP](https://timoni.sh/ai-agents)
can help look up authoring and lifecycle details in VS Code; they do not
replace rendering, review or kro validation. This checkout does not install
editor extensions or agent integrations.

## Additional Timoni features to evaluate

Beyond this RGD module, the [official feature index](https://timoni.sh/llms.txt)
covers Kubernetes/CRD schema vendoring, importing existing YAML, embedded
files, immutable ConfigMaps/Secrets, Kubernetes version constraints,
multi-step apply, custom health checks and post-deploy Job tests. Modules
support OCI versioning/digest pinning and signature verification workflows.
Bundles compose module instances; runtimes inject Kubernetes/CI values and
support multi-cluster configuration. Secrets can be supplied through
Kubernetes, CI, SOPS or 1Password workflows. Bundle update policies, GitHub
Actions and agent tooling help automate authoring and releases.

These features are separate from kro's graph execution. For a first company
pilot, evaluate the concrete RGD authoring/review loop and Helm-free release
boundary first; add Flux, runtime injection or multi-cluster delivery only
when required. See [the delivery guide](lifecycle.md) for the distinction
between direct Timoni lifecycle and Flux-managed rendered manifests.

## Hands-on track

1. [Author a real RGD](exercises/01-translate-rgd-patterns.md).
2. [Package and reason about application lifecycle](exercises/02-lifecycle-oci-flux.md).
3. Compare the result with the existing [kro workshop](../workshop/README.md).
