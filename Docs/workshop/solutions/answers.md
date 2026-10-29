# Workshop self-checks

These are reference points, not the only acceptable answers. Use the current
RGDs as the source of truth; avoid treating a live observation during
reconciliation as a transactional guarantee.

## 01 - Observe the graph

- The user-facing object is a `Dungeon` custom resource. Its `spec` holds
  requested configuration and saved game fields; its `status` summarizes
  derived information such as living monster count, boss state/phase, victory,
  and treasure state.
- The Dungeon RGD directly composes Hero, Monster (one per `monsterHP` entry),
  Boss, Treasure, and a configuration ConfigMap. The Hero, Monster, Boss,
  Treasure, and Loot RGDs compose supporting ConfigMaps and, conditionally,
  Secrets or Loot resources.
- Labels, predictable names, status fields, and printer columns help operators
  identify and inspect instances. They improve observability but are not by
  themselves an authorization boundary.
- A turn first changes desired game fields written by the Go API. kro then
  projects those fields into the composed resources and derives status. These
  are asynchronous reconciliations, not a multi-resource atomic transaction;
  readers should tolerate temporary lag.
- The Monster graph includes Loot when `schema.spec.hp == 0`; the Treasure
  graph includes a Secret when `schema.spec.opened == 1`. The Dungeon is a
  useful grouping and lifecycle boundary in this demo, but production ownership,
  authorization, and cleanup policy still need explicit design.
- Upstream kro v0.9.4 tracks composed resources using ApplySet inventory and
  kro-managed metadata/finalizers. Do not assume the demo's children are
  cleaned up through ordinary owner-reference garbage collection.

## 02 - Read RGDs and CEL

- `dungeon-graph.yaml` defines `Dungeon` in `game.k8s.example`. Its schema
  declares defaults, allowed values, and bounds; these form part of the API
  contract. Printer columns surface selected status/spec fields to kubectl
  users.
- `forEach` uses `lists.range(size(schema.spec.game.monsterHP))`, so one
  Monster is composed per array entry, with a stable zero-based index.
- The Dungeon RGD's `readyWhen` checks
  `${hero.status.maxHP == '200'}`. That is a composed-resource readiness
  condition; it is not the game's "boss unlocked" rule. Game progression is
  separately reflected in Boss state derived from `monstersAlive`.
- The Monster RGD includes its Loot resource when HP is zero. The Treasure RGD
  includes its Secret when `opened == 1`. A condition controls whether the
  resource is part of the graph; it is different from merely writing an
  "unopened" status value to an always-present ConfigMap.
- The Boss RGD classifies phase1 while HP is greater than 50% of maximum,
  phase2 at or below 50% and above 25%, and phase3 at or below 25% while HP
  remains positive; zero HP is defeated. Its exact nested condition checks
  zero first when deriving entity state and phase.
- `&&` ensures the victory summary requires both a defeated boss and a living
  hero. The CEL playground expressions read the current saved Dungeon data
  and do not change it. RGD expressions can also reference other graph
  resources, unlike the playground's saved Dungeon data context.
- ConfigMap data is string-valued. The RGD explicitly uses `string(...)` or
  string-valued expressions where appropriate; type awareness matters when
  composing and reading values.
- On a turn, Go computes game rules and writes `Dungeon.spec.game`. kro
  reconciles children and CEL-derived projections. The UI playground is a
  read-only CEL evaluator, not the controller's expression runner.

## 03 - Platform design review

A technical comparison can:

- select a recurring composition problem with enough repetition to justify an
  API abstraction;
- define a narrow, comprehensible interface with safe defaults and useful
  constraints;
- identify ownership of API evolution, controller failures,
  cleanup, documentation, permissions, and support;
- account for tenant isolation, resource quotas, secret handling, observability,
  version compatibility, and rollback;
- compare kro with existing mechanisms rather than introducing a new platform
  dependency by default; and
- distinguish this game's desired-state/reconciliation demonstration from a
  production workload's data storage or authentication architecture.

The game demonstrates composition, generated custom APIs, CEL projections,
collection and conditional resources, readiness, and reconciliation. It does
not demonstrate kro's internal implementation, the original fork's
state-write combat engine, production multi-tenant security, or a high-throughput
business data store.
