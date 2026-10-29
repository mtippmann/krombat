# Krombat: author kro RGDs in CUE, package them with Timoni

This module renders the **six real kro ResourceGraphDefinitions**, not a
snapshot of the resources kro would create. kro still generates the
Dungeon/Hero/Monster/Boss/Treasure/Loot APIs and reconciles their instances.
Go still computes turns. Timoni packages and delivers the graph definitions.
No Helm chart is involved.

Start with [`templates/monster.cue`](templates/monster.cue) beside
[`manifests/rgds/monster-graph.yaml`](../../manifests/rgds/monster-graph.yaml),
then read the [walkthrough](../../Docs/timoni/rgd-mapping.md).

## Offline commands

Use Timoni 0.35.0, which embeds CUE 0.17.1. From the repository root:

```bash
timoni fmt --diff timoni/krombat
timoni mod vet timoni/krombat --name krombat-graphs --namespace rpg-system
timoni build krombat-graphs ./timoni/krombat --namespace rpg-system
timoni build krombat-graphs ./timoni/krombat --namespace rpg-system \
  --values timoni/krombat/debug_values.cue
```

Both builds emit six cluster-scoped RGDs. They do not emit live game
instances, their child resources, the game Deployments, or the kro controller.
`rpg-system` is the namespace a future Timoni instance inventory would use;
the RGD objects themselves do not have a namespace.

To validate default/override configurations, compare all graph contents with
the current YAML, and check local OCI packaging:

```bash
bash timoni/test.sh
```

The comparison tests require Python 3 and PyYAML 6.x in addition to Timoni.
All checks are offline and do not read kubeconfig or contact a cluster.

## What CUE simplifies

| File | Responsibility |
|---|---|
| `templates/common.cue` | Shared RGD envelope, child metadata, state ConfigMaps, child CR templates, conditional Loot and printer columns |
| `templates/{hero,monster,boss,treasure,loot,dungeon}.cue` | The actual six graph definitions, retaining kro SimpleSchema and CEL |
| `templates/config.cue` | Typed module values and graph assembly |
| `timoni.cue` | Emit those graphs as the Timoni module's managed resources |
| `debug_values.cue` | Example platform label and new-Dungeon API default overrides |
| `cue.mod/gen/kro.run/` | RGD validation schema generated from upstream kro v0.9.4 |

The graph objects are constrained by the vendored upstream RGD schema. Shared
CUE definitions reduce repeated metadata and resource templates. `${...}`
remains a literal string for kro to evaluate at reconciliation time.
`includeWhen` and `forEach` remain fields in the generated RGD, not CUE
conditions or loops over current game data.

Default output matches `manifests/rgds/*-graph.yaml` structurally, with
`spec.schema.scope: Namespaced` made explicit (the upstream API default).
The active YAML remains unchanged and is still used by `scripts/local.sh`.

## Values are platform release configuration

```cue
values: {
	graphLabels: "platform.example/team": "workshop"
	dungeonDefaults: {
		monsters:   5
		difficulty: "normal"
	}
}
```

This configures the RGD release: its metadata labels and the generated
Dungeon API's defaults. It does not set any Dungeon's current HP or change
turn calculations. API defaults apply when fields are omitted; the existing
Go API can explicitly supply its own values.

## Local OCI artifact

Choose a new output path; Timoni rejects existing outputs:

```bash
timoni mod build ./timoni/krombat --version 0.1.0 \
  --output /tmp/krombat-rgds-0.1.0.oci.tar
```

This creates a local archive, not a registry push. Do not apply this module
over the existing six graphs: they have cluster-wide names and are already
managed by the local demo workflow. Another namespace or Timoni instance name
does not isolate them. See [lifecycle and delivery](../../Docs/timoni/lifecycle.md)
for a Helm-free platform/application release design and ownership constraints.
