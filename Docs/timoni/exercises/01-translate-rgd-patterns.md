# 01 - Maintain a real RGD with CUE

**Goal:** Reduce graph-authoring repetition while preserving the actual RGD
and its runtime semantics. All commands below are offline.

## Read and render

Open the original
[`monster-graph.yaml`](../../../manifests/rgds/monster-graph.yaml) beside
[`monster.cue`](../../../timoni/krombat/templates/monster.cue) and
[`common.cue`](../../../timoni/krombat/templates/common.cue).
Read the [walkthrough](../rgd-mapping.md), then run from the repository root:

```bash
timoni build krombat-graphs ./timoni/krombat --namespace rpg-system
bash timoni/test.sh
```

Find `kind: ResourceGraphDefinition`, `monsterState`, `loot`, `includeWhen`
and the original CEL in the output. Find the Dungeon graph's `forEach` and
status projection. Nothing has expanded the Monster collection yet: the
generated object is a graph definition for kro to execute.

## Trace shared structure

1. How does `#RGD` supply the common API group/version and RGD envelope?
2. Which fields does `#StateResource` remove from each handwritten state
   template?
3. How do Monster and Boss share `#LootResource` while preserving their
   different rarity rules?
4. Which concrete strings are SimpleSchema, and which are CEL? Who evaluates
   each, and when?
5. Why would using CUE `if hp == 0` instead of emitting `includeWhen` change
   the layer that owns runtime behavior?

## Change release values, not gameplay state

Render with the existing example values:

```bash
timoni build krombat-graphs ./timoni/krombat --namespace rpg-system \
  --values timoni/krombat/debug_values.cue
```

Compare with the default output. Only the RGD labels and the Dungeon API's
default monster count/difficulty should change; all six graphs, dependencies,
CEL and resource templates remain.

Temporarily change `dungeonDefaults.monsters` to `0` in `debug_values.cue`:

```bash
timoni mod vet timoni/krombat --debug
```

Observe the constraint error and restore `5`. This validates a module
authoring input; it is not a call to the Dungeon API.

## Review a graph change

On a temporary edit, change the shared `#LootResource` name suffix in
`common.cue`. Render again and identify **both** affected graphs. The parity
test should fail: it protects exact compatibility with the original YAML,
including CEL strings. Restore the original suffix and rerun the checks.

Do not apply the generated graphs or update the reference YAML as part of
this exercise. In a real platform change, an intentional generated-YAML diff,
API compatibility review and kro-level validation would be required.

## Deliverable

Explain one duplication removed by CUE and one behavior that remains entirely
owned by kro. Show a generated RGD diff rather than a precomputed ConfigMap
snapshot.
