# Local operations

Every command targets minikube explicitly. Never change the shared current
context, use a cloud cluster, or remove the shared kro installation.

## Check

```bash
bash scripts/local.sh status
getent ahostsv4 krombat.test
minikube -p minikube ip
curl --fail http://krombat.test/
kubectl --context minikube -n rpg-system get pods
kubectl --context minikube -n rpg-system logs deployment/rpg-backend --tail=50
kubectl --context minikube -n kro-system logs deployment/kro --tail=50
kubectl --context minikube -n rpg-demo get events --sort-by=.lastTimestamp
```

An nginx 404 before deployment means DNS and ingress connectivity work, but no
matching application route exists. After deployment, inspect the `krombat`
Ingress and service endpoints for routing problems. A 502/503 usually means the
backend or frontend has not become ready.

```bash
kubectl --context minikube -n rpg-system describe ingress krombat
kubectl --context minikube -n rpg-system get endpointslices
kubectl --context minikube -n rpg-demo describe dungeon team-demo
kubectl --context minikube get rgd dungeon-graph -o yaml
```

The host's `krombat.test` mapping is already configured; deployment does not
edit resolver configuration or hosts files. If it regresses, compare the host
lookup with the minikube IP before making separately approved DNS changes.

## Rebuild and restart

```bash
bash scripts/local.sh build
bash scripts/local.sh deploy
```

Images are explicitly loaded into the minikube profile and use
`imagePullPolicy: Never`. The deploy command restarts only the two game
deployments so a reused local image tag is picked up.

The deploy helper checks the existing controller image, ownership of its
namespaces/graphs/CRDs, dependency ordering, graph readiness, app rollouts,
and ingress response. It generates the session Secret only when missing.
No Secret value is written into the repository.

## Reset a run

Prefer **Delete dungeon** in the UI, then create a fresh named dungeon.
For the documented presentation instance only:

```bash
kubectl --context minikube -n rpg-demo delete dungeon team-demo \
  --ignore-not-found --wait=true
kubectl --context minikube -n rpg-demo get hero,monster,boss,treasure,loot,cm,secret \
  -l game.k8s.example/dungeon=team-demo
```

Do not strip finalizers to force deletion. Inspect graph conditions and kro
logs if inventory cleanup stalls. Removing a Namespace is not a normal reset.

## Ownership and drift

kro v0.9.4 records managed resources using ApplySet inventory and labels. Its
finalizer performs ordered pruning. Parent-child identities are available in
`kro.run/instance-id`, not necessarily `metadata.ownerReferences`.

```bash
kubectl --context minikube -n rpg-demo get cm team-demo-hero -o yaml
kubectl --context minikube -n rpg-demo patch cm team-demo-hero --type=merge \
  -p '{"data":{"hp":"9999"}}'
kubectl --context minikube -n rpg-demo get cm team-demo-hero -w
```

Stop the watch after the real hero HP returns. Only edit this explicitly
demo-owned resource, not another workload's configuration.

## Remove this demo

First delete each intended Dungeon **by name** and wait for its children to
disappear. Confirm there are no runs you want to preserve:

```bash
kubectl --context minikube -n rpg-demo get dungeons
```

After named runs have been removed, remove only the app/RBAC manifests and the
six graphs installed by this checkout:

```bash
kubectl --context minikube delete -f manifests/system --ignore-not-found
kubectl --context minikube delete -f manifests/rbac/rbac.yaml --ignore-not-found
kubectl --context minikube delete -f manifests/rgds --ignore-not-found
kubectl --context minikube -n rpg-system delete secret krombat-session --ignore-not-found
```

Retain the two namespaces unless you have inspected their contents and explicitly
decided to remove them. Do not uninstall kro, nginx, monitoring, the existing
`my-application` graph, or the minikube profile. Graph deletion removes generated
API definitions; never do this while preserving game instances.
