# Kustomize and CUE: replica comparison

This offline example accompanies slide 5. It uses the same Deployment as
the [Timoni example](../example/README.md), with a base replica count of 2.
Two Kustomize overlays transform the count to 3 or 0.

From `Docs/presentation`:

```bash
kubectl --context=minikube kustomize example-kustomize/base
kubectl --context=minikube kustomize example-kustomize/three
kubectl --context=minikube kustomize example-kustomize/zero

timoni build demo ./example -n demo
timoni build demo ./example -n demo --values ./example/three.yaml
timoni build demo ./example -n demo --values ./example/invalid.yaml
```

These are local rendering commands; they do not contact Kubernetes. The
explicit `minikube` context follows this checkout's command convention.
It does not change the current kubeconfig context. `docker-desktop` is the
other permitted context.

| Input | Kustomize | Timoni / CUE |
|---|---|---|
| Base / no input | Render 2 replicas | Render the default of 2 |
| 3 | Render 3 replicas | Render 3 replicas |
| 0 | Render 0 replicas | Fail: outside the module's 1–10 range |

The first two rendered Deployments are structurally identical between
the tools. Zero is a valid Kubernetes replica count, but it violates the
example CUE module's declared configuration contract. This comparison
shows where the rule lives; it does not imply that Kustomize cannot be
paired with schema, policy or CI validation.

Checked with kubectl 1.37.1 / embedded Kustomize 5.8.1 and Timoni 0.35.0 /
embedded CUE 0.17.1. No apply, deployment or live reconciliation is part
of this example.
