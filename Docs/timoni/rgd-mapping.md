# YAML to CUE: author the RGD itself

**CUE is an authoring language here, not a replacement graph engine.**
The output must still be `kind: ResourceGraphDefinition`. The earlier
snapshot-style example modeled the wrong layer; it has been replaced.

## 1. Start with the actual Monster RGD

Read [`manifests/rgds/monster-graph.yaml`](../../manifests/rgds/monster-graph.yaml).
Its resource list contains a state ConfigMap and conditional Loot CR:

```yaml
resources:
  - id: monsterState
    template:
      apiVersion: v1
      kind: ConfigMap
      metadata:
        name: ${schema.metadata.name}
        namespace: ${schema.metadata.namespace}
        labels:
          game.k8s.example/dungeon: ${schema.spec.dungeonName}
          game.k8s.example/entity: monster
      data:
        entityState: "${schema.spec.hp > 0 ? 'alive' : 'dead'}"
        hp: ${string(schema.spec.hp)}
  - id: loot
    includeWhen:
      - ${schema.spec.hp == 0}
    template:
      apiVersion: game.k8s.example/v1alpha1
      kind: Loot
      metadata:
        name: ${schema.metadata.name + '-loot'}
        namespace: ${schema.metadata.namespace}
        labels:
          game.k8s.example/dungeon: ${schema.spec.dungeonName}
      spec:
        dungeonName: ${schema.spec.dungeonName}
        rarity: "${schema.spec.index % 2 == 0 ? 'rare' : 'common'}"
```

Boss has almost the same ConfigMap envelope and defeated-entity Loot
template. Across the six files, metadata, graph headers and printer columns
also repeat. These are the authoring/maintenance opportunities.

## 2. Share the structure, retain graph-specific CEL

The real [`templates/monster.cue`](../../timoni/krombat/templates/monster.cue)
defines:

```cue
#MonsterGraph: #RGD & {
	#name: "monster-graph"
	#kind: "Monster"
	spec: {
		schema: {
			additionalPrinterColumns: [#HPColumn, #StateColumn, #ReadyColumn]
			spec: {
				dungeonName: "string | required=true"
				index:       "integer | required=true minimum=0"
				hp:          "integer | required=true minimum=0"
			}
			status: {
				entityState: "${monsterState.data.entityState}"
				hp:          "${monsterState.data.hp}"
			}
		}
		resources: [
			#StateResource & {
				id:      "monsterState"
				#entity: "monster"
				template: data: {
					entityState: "${schema.spec.hp > 0 ? 'alive' : 'dead'}"
					hp:          "${string(schema.spec.hp)}"
				}
			},
			#LootResource & {
				template: spec: rarity: "${schema.spec.index % 2 == 0 ? 'rare' : 'common'}"
			},
		]
	}
}
```

`&` unifies compatible structures/constraints. It is not a textual include or
last-wins override: contradictory concrete values fail. `#` definitions and
their helper inputs do not become extra fields in the emitted Kubernetes
object.

The helpers live in
[`templates/common.cue`](../../timoni/krombat/templates/common.cue):

- `#RGD` adds the `ResourceGraphDefinition` envelope, fixed API identity,
  platform labels and instance scope. It is constrained by the vendored
  upstream `rgdv1.#ResourceGraphDefinition`.
- `#StateResource` supplies the ConfigMap's API version, kind, child metadata
  and entity label.
- `#LootResource` supplies the Loot CR template, name, Dungeon reference and
  runtime inclusion condition. Its definition is:

```cue
#LootResource: #GameResource & {
	#kind: "Loot"
	#name: "${schema.metadata.name + '-loot'}"
	id:    "loot"
	includeWhen: ["${schema.spec.hp == 0}"]
}
```

The same helper is used in
[`templates/boss.cue`](../../timoni/krombat/templates/boss.cue), but the Boss
sets `template: spec: rarity: "epic"`. Fix the shared Loot naming or metadata
once, rather than editing two copied templates. The graphs keep their
different state/phase expressions where reviewers can see them.

## 3. Render it: it is still a kro RGD

```bash
timoni build krombat-graphs ./timoni/krombat --namespace rpg-system
```

The output includes:

```yaml
apiVersion: kro.run/v1alpha1
kind: ResourceGraphDefinition
metadata:
  name: monster-graph
  labels:
    app.kubernetes.io/part-of: krombat
spec:
  schema:
    apiVersion: v1alpha1
    group: game.k8s.example
    kind: Monster
    scope: Namespaced
    # The original schema, status and printer columns follow.
  resources:
    # The original expanded monsterState and loot templates follow.
```

The full output has the original `${...}` expressions and `includeWhen`.
No Monster is created at build time. After a separately approved deployment,
kro receives this RGD, generates the Monster API and evaluates it for each
Monster instance as before.

## 4. Keep SimpleSchema and CEL in their own layers

| Source in CUE | What CUE does | What kro does later |
|---|---|---|
| `hp: "integer \| required=true minimum=0"` | Emits a SimpleSchema string inside the RGD | Builds the generated API schema |
| `includeWhen: ["${schema.spec.hp == 0}"]` | Emits the condition unchanged | Re-evaluates it when reconciling; includes/prunes Loot |
| `forEach: [{idx: "${lists.range(size(schema.spec.game.monsterHP))}"}]` | Emits one graph collection definition | Expands it into the current number of Monsters |
| `readyWhen: ["${hero.status.maxHP == '200'}"]` | Emits the readiness rule | Evaluates resource readiness in the graph |
| `status: bossState: "${boss.status.entityState}"` | Emits a projection expression | Observes the child and projects live Dungeon status |
| `trophy: "${random.seededString(8, schema.metadata.name)}"` | Emits the function call unchanged | Evaluates its deterministic trophy expression |

Those are **not** CUE `if`/`for` operations on HP snapshots. In particular,
the Dungeon graph's `forEach` stays in
[`templates/dungeon.cue`](../../timoni/krombat/templates/dungeon.cue).
Neither Timoni nor Flux manages the Hero/Monster/etc. objects that kro creates.

There are two distinct validations: CUE/Timoni validates module values and
the RGD object's outer schema; kro validates/compiles its embedded graph and
CEL. The RGD's `template` and SimpleSchema fields are intentionally flexible.
Do not claim editor completion or `mod vet` proves that every nested template,
CEL reference, dependency or runtime condition is correct.

## 5. Use typed values for deliberate API-release choices

[`templates/config.cue`](../../timoni/krombat/templates/config.cue) exposes:

```cue
#Config: {
	graphLabels: timoniv1.#Labels
	dungeonDefaults: {
		monsters:   *3 | int & >=1 & <=10
		difficulty: *"easy" | "easy" | "normal" | "hard"
	}
}
```

[`templates/dungeon.cue`](../../timoni/krombat/templates/dungeon.cue) turns
these into the RGD's SimpleSchema strings:

```cue
monsters:   "integer | default=\(#config.dungeonDefaults.monsters) minimum=1 maximum=10"
difficulty: "string | default=\"\(#config.dungeonDefaults.difficulty)\" enum=easy,normal,hard"
```

`\(...)` is **CUE interpolation**, evaluated while building the module.
`${...}` is **kro CEL**, preserved for reconciliation. Keeping the two
notations visible makes their responsibilities explicit.

The [debug values](../../timoni/krombat/debug_values.cue) change defaults to
5 and normal and add a platform label:

```bash
timoni build krombat-graphs ./timoni/krombat --namespace rpg-system \
  --values timoni/krombat/debug_values.cue
```

This changes the graph/API contract, not an active game's HP. Existing
objects with explicit fields are not reset; clients that supply explicit
creation values need not follow the API defaults. Treat default changes as
reviewed API changes, not per-player difficulty selection.

## 6. Check the maintenance invariant

```bash
bash timoni/test.sh
```

It compares every field of all six default-rendered RGDs against the original
YAML, including CEL strings, types and resource-list order. It only normalizes
the original implicit `Namespaced` scope. It also checks that overrides only
change the intended defaults/labels, invalid values fail, and OCI packaging
works locally. The existing YAML remains the deployment baseline.

This provides evidence that introducing an authoring layer did not change
the graph. Future intentional graph changes should be reviewed as generated
RGD diffs, then validated with kro in an explicitly approved disposable
environment. This track does not apply the module to the shared demo.

## Creating another graph

Use `hero.cue` as the smallest starting point: declare a new `#...Graph`
unified with `#RGD`, choose a distinct graph name and generated kind, and
write its SimpleSchema, status and resource list. Reuse `#StateResource` only
if the new API has the same Dungeon reference and state-ConfigMap pattern;
otherwise compose a `#Resource` with its own template.

Register the graph in `#Instance.graphs` in `config.cue` and the explicit
emission list in `timoni.cue`. Render before publishing. Adding a new RGD
introduces a new cluster API; it should be a deliberate platform change,
not an incidental consequence of a package loop. The default parity test
intentionally expects the current six APIs, so a planned expansion also
requires a reviewed baseline/test update.
