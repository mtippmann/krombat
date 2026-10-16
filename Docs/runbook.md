# Local operations

The examples below target minikube explicitly. For Docker Desktop, set
`KUBECTL_CONTEXT=docker-desktop` for scripts and use `--context docker-desktop`
instead of `--context minikube` in manual commands. Never change the shared
current context, use a cloud cluster, or remove the shared kro installation.

## Docker Desktop on Windows

Run scripts from Git Bash. Follow the [Windows setup](../README.md#windows-docker-desktop)
for the explicit setup command and manual Windows hosts-file entry.
Setup installs only missing controllers; it refuses namespace collisions and
never upgrades existing installations.

No public image push is needed: kubeadm uses Desktop's local images and kind
nodes receive an explicit image import. If a kind node cannot be inspected,
enable **Show system containers (advanced)**. Use the containerd image store
for kind. Ensure your Docker CLI targets Docker Desktop with Linux containers,
not a remote daemon or a minikube Docker environment.

For `ErrImageNeverPull`, rerun `build` against the correct Docker daemon and
then `deploy`. After switching Docker image stores, rebuild: images in the
previous store are hidden. For `ImagePullBackOff`, inspect which pod failed:
application images never pull; controller images need internet access.

If localhost port 80 is unavailable, inspect the ingress LoadBalancer service
and any Windows process already binding that port. Setup checks HTTP connectivity
but cannot identify an unrelated HTTP server on the same port; deployment's
application readiness checks are still required. Do not change the game URL or
expose the demo publicly.

Browser containers map `krombat.test` to Docker's host gateway. If host access
works but browser tests fail, check Docker Desktop networking and Windows
firewall permissions for the ingress endpoint. Host networking is not required.

## Check

```bash
bash scripts/local.sh status
getent ahostsv4 krombat.test
minikube -p minikube ip
curl --fail http://krombat.test/
kubectl -n rpg-system get pods
kubectl -n rpg-system logs deployment/rpg-backend --tail=50
kubectl -n kro-system logs deployment/kro --tail=50
kubectl -n rpg-demo get events --sort-by=.lastTimestamp
```

An nginx 404 before deployment means DNS and ingress connectivity work, but no
matching application route exists. After deployment, inspect the `krombat`
Ingress and service endpoints for routing problems. A 502/503 usually means the
backend or frontend has not become ready.

```bash
kubectl -n rpg-system describe ingress krombat
kubectl -n rpg-system get endpointslices
kubectl -n rpg-demo describe dungeon team-demo
kubectl get rgd dungeon-graph -o yaml
```

The host's `krombat.test` mapping is already configured; deployment does not
edit resolver configuration or hosts files. If it regresses, compare the host
lookup with the minikube IP before making separately approved DNS changes.

## Rebuild and restart

```bash
bash scripts/local.sh build
bash scripts/local.sh deploy
```

The deploy helper checks the existing controller image, ownership of its
namespaces/graphs/CRDs, dependency ordering, graph readiness, app rollouts,
and ingress response. It generates the session Secret only when missing.
No Secret value is written into the repository.

## Reset a run

Prefer **Delete dungeon** in the UI, then create a fresh named dungeon.
For the documented presentation instance only:

```bash
kubectl -n rpg-demo delete dungeon team-demo \
  --ignore-not-found --wait=true
kubectl -n rpg-demo get hero,monster,boss,treasure,loot,cm,secret \
  -l game.k8s.example/dungeon=team-demo
```

Do not strip finalizers to force deletion. Inspect graph conditions and kro
logs if inventory cleanup stalls. Removing a Namespace is not a normal reset.

## Ownership and drift

kro v0.9.4 records managed resources using ApplySet inventory and labels. Its
finalizer performs ordered pruning. Parent-child identities are available in
`kro.run/instance-id`, not necessarily `metadata.ownerReferences`.

```bash
kubectl -n rpg-demo get cm team-demo-hero -o yaml
kubectl -n rpg-demo patch cm team-demo-hero --type=merge \
  -p '{"data":{"hp":"9999"}}'
kubectl -n rpg-demo get cm team-demo-hero -w
```

Stop the watch after the real hero HP returns. Only edit this explicitly
demo-owned resource, not another workload's configuration.

## Remove this demo

First delete each intended Dungeon **by name** and wait for its children to
disappear. Confirm there are no runs you want to preserve:

```bash
kubectl -n rpg-demo get dungeons
```

After named runs have been removed, remove only the app/RBAC manifests and the
six graphs installed by this checkout:

```bash
kubectl delete -f manifests/system --ignore-not-found
kubectl delete -f manifests/rbac/rbac.yaml --ignore-not-found
kubectl delete -f manifests/rgds --ignore-not-found
kubectl -n rpg-system delete secret krombat-session --ignore-not-found
```

Retain the two namespaces unless you have inspected their contents and explicitly
decided to remove them. Do not uninstall kro, nginx, monitoring, the existing
`my-application` graph, or the minikube profile. Graph deletion removes generated
API definitions; never do this while preserving game instances.
