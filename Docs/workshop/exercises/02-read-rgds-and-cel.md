# 02 - Read RGDs and CEL

**Goal:** Read the current graph definitions as a platform consumer or
maintainer, and distinguish an RGD controller expression from the demo's
read-only CEL playground.

Start with:

- `manifests/rgds/dungeon-graph.yaml`
- `manifests/rgds/monster-graph.yaml`
- `manifests/rgds/boss-graph.yaml`
- `manifests/rgds/treasure-graph.yaml`
- `manifests/rgds/loot-graph.yaml`

The game UI's **CEL playground** evaluates expressions against the latest
Dungeon's saved `schema.spec.game` data. It does not mutate the Dungeon. The
evaluation context is not identical to the RGD controller context: an RGD may
also refer to graph resources such as `hero`, `monsters`, or `config`.

## Activity A - Schema and API contract

1. In `dungeon-graph.yaml`, find the RGD schema for `Dungeon`.
2. Identify a default, an enum, and a numeric bound.
3. Find the `additionalPrinterColumns` and the graph's `status` fields.

### Questions

1. What API group and kind does this RGD define?
2. Which values constrain Dungeon creation, and which values are observed
   outputs?
3. Why are schema defaults and bounds useful in a platform API? What
   assumptions or compatibility concerns would arise if the platform team
   changed one?
4. Which status fields summarize other graph resources rather than simply
   echoing input?

## Activity B - Resource composition patterns

Inspect the `resources` sections in the Dungeon, Monster, Boss, and Treasure
RGDs.

### Questions

1. In the Dungeon graph, how does `forEach` derive the number of Monster
   resources and their indices from `schema.spec.game.monsterHP`?
2. Which graph uses `readyWhen`? State the exact condition. Is it a business
   readiness condition such as "the boss is unlocked," or a controller
   readiness check on a composed resource?
3. Find an `includeWhen` condition in the Monster graph and one in the
   Treasure graph. Which object is conditional in each case?
4. When the Boss reaches zero HP, what changes besides its derived state?
5. The Treasure graph always composes a state ConfigMap, but conditionally
   composes a Secret. What design distinction does that illustrate?

## Activity C - Evaluate CEL

In the game, open the CEL playground for your Dungeon and evaluate these
read-only expressions one at a time:

```cel
schema.spec.game.monsterHP.filter(hp, hp > 0).size()
schema.spec.game.bossHP == 0
schema.spec.game.treasureOpened == 1
schema.spec.game.heroHP
```

Then compare your results to the expressions and templates in the RGDs.

### Questions

1. What does each expression return in the state you observed?
2. How does the first expression use a collection operation and a predicate?
3. In the Boss RGD, explain how the nested conditional expressions classify
   phases using HP relative to `maxHP`. At which boundaries do phase2 and
   phase3 begin?
4. In the Dungeon RGD, why is
   `schema.spec.game.bossHP == 0 && schema.spec.game.heroHP > 0` a more precise
   victory summary than checking only whether the boss is defeated?
5. What do you need to check when reading an expression for types? For
   example, the graph converts some numbers to strings when writing
   ConfigMap data.

## Activity D - Trace ownership and operational meaning

Use the labels and resource templates to answer:

1. How are child objects named and grouped with their Dungeon?
2. What evidence helps you identify kro-managed resources in this demo?
3. Why should you not assume the children use ordinary `ownerReferences` and
   Kubernetes garbage collection for cleanup?
4. What should a client do if the root spec has changed but a projected status
   or child has not converged yet?

## Deliverable

Annotate one RGD with short notes for its schema, each resource, one CEL
expression, readiness, and status. Do not change or apply the RGD.
