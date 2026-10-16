# Team presentation: a game as a resource graph

## Prepare

For Windows Git Bash with Docker Desktop, follow the
[setup instructions](../../README.md#windows-docker-desktop), then export
`KUBECTL_CONTEXT=docker-desktop` before running scripts. Replace
`--context minikube` with `--context docker-desktop` in the manual examples
below; never switch the shared current context.

```bash
bash scripts/local.sh build
bash scripts/local.sh deploy
bash tests/run.sh
```

Open `http://krombat.test` and log in with **Enter local demo**.
Use three monsters, easy difficulty and the name `team-demo`.
If that run already exists, delete it through the UI and recreate it.
Alternatively apply the fully initialized example:

```bash
kubectl --context minikube apply -f Docs/demo/dungeon-demo.yaml
```

Do not reapply that example over a live run: it resets desired HP fields.

## Set up your inspection tool

```bash
k9s --context minikube -n rpg-demo
```

In k9s use `:dungeons`, `:monsters`, `:bosses`, `:loots`, `:configmaps`,
and `:secrets`; inspect YAML and events. Dungeon printer columns show living
monsters, boss state, phase, hero HP, and readiness.

For Radar, use a **minikube-only kubeconfig** so no shared context can be switched:

```bash
umask 077
RADAR_CONFIG=$(mktemp)
kubectl --context minikube config view --minify --flatten --raw > "$RADAR_CONFIG"
radar --kubeconfig "$RADAR_CONFIG" --listen-address 127.0.0.1
# After stopping Radar:
rm "$RADAR_CONFIG"
```

The temporary file contains cluster credentials; never print or commit it.
Verified with Radar v1.16.0: open **Resources**, expand **Example Game**, choose
**Dungeon**, and click `team-demo`. The list shows the custom Monsters/Boss/Phase/
HeroHP/Ready columns; the detail drawer shows live status, conditions, spec.game,
labels, ApplySet annotations and YAML. The direct list path is
`/resources/dungeons?apiGroup=game.k8s.example`.

Radar's custom-resource views can inspect these objects. Its automatic topology
may not infer every arbitrary CR relationship: upstream kro uses ApplySet labels
rather than default ownerReferences. Use the game's live resource inspector,
Radar's resource details and k9s/YAML as reliable complementary views.

## Walkthrough

1. **A new Kubernetes API.** Show `dungeon-graph`, its schema and a generated
   Dungeon CRD. Explain that schema validation, RBAC, status and printer columns
   are native Kubernetes facilities.
2. **One input, many resources.** Create the dungeon through the UI and inspect
   Hero, three Monsters, Boss, Treasure and ConfigMaps. Show `forEach` and the
   composition of six independent RGDs. Compare to an application stack spawning
   databases, services and monitoring.
3. **A turn is a desired-state update.** Attack a monster. Go writes spec.game;
   kro updates the Monster CR and ConfigMap. The UI inspector shows the real
   objects. Explain asynchronous convergence rather than inventing a synchronous
   Kubernetes transaction across all children.
4. **Conditional resources.** Kill a monster and inspect its newly created Loot
   CR, ConfigMap and Secret. Show `includeWhen`. These are trophies, not an
   equipment subsystem. Compare to optional backups or add-on infrastructure.
5. **Derived state and aggregation.** Clear monsters, observe the boss becoming
   ready, then attack until phase2 and phase3. Show the 50%/25% CEL expressions.
   Explain the distinction between gameplay readiness and controller readiness.
6. **Treasure.** Defeat the boss, open the chest and inspect the conditional
   Treasure Secret. Evaluate `schema.spec.game.bossHP == 0` in the CEL playground.
7. **Drift repair.** Edit `team-demo-hero` ConfigMap hp to `9999` and observe kro
   restore the real value. Delete that same ConfigMap and watch it return with
   a new UID. Do not change any unrelated resource.
8. **Lifecycle.** Delete the dungeon through the UI and observe children vanish.
   Explain ApplySet inventory, the `kro.run/finalizer` and ordered cleanup.

Useful read-only commands:

```bash
kubectl --context minikube get rgd -l app.kubernetes.io/part-of=krombat
kubectl --context minikube get crd dungeons.game.k8s.example -o yaml
kubectl --context minikube -n rpg-demo get dungeon team-demo -o yaml
kubectl --context minikube -n rpg-demo get hero,monster,boss,treasure,loot,cm,secret \
  -l game.k8s.example/dungeon=team-demo
kubectl --context minikube -n rpg-demo get cm team-demo-hero -o yaml
```

## Be explicit about the limits

This is Kubernetes orchestration applied to a fun domain, not a recommendation
to use Kubernetes as a high-throughput game database. Go calculates turn damage
and validates actions; stock kro projects desired resources and evaluates CEL
for derived state. The original game's custom write-back engine, AWS deployment
and additional gameplay are not present.

The shared local identity and unencrypted HTTP are intentional for this machine.
Do not present this as a production authentication design.
