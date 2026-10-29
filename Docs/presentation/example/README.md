# Deployment example for the Timoni and CUE introduction

This small, self-contained Timoni module supports slides 3–5. It
renders one Deployment. Its single configurable value is `replicas`:

```cue
#Config: {
    replicas: *2 | (int & >=1 & <=10)
}
```

The default is 2; valid overrides are integers from 1 to 10. The module
uses the value for the Deployment's `spec.replicas`. The example keeps all
its CUE definitions in [timoni.cue](timoni.cue) to make the connection easy
to follow. It omits vendored Kubernetes schemas to keep the example small;
production modules can use vendored API schemas and additional validation.

## Try it offline

From this directory, with the Timoni CLI installed:

```bash
# Default: the output contains spec.replicas: 2.
timoni build demo . --namespace demo

# Valid override: the output contains spec.replicas: 3.
timoni build demo . --namespace demo --values three.yaml

# Invalid override: the command exits with an error for replicas: 0.
timoni build demo . --namespace demo --values invalid.yaml
```

Both values files have a top-level `values` key. `timoni build` emits YAML;
it does not apply the Deployment. The sample image and Deployment are for
learning about module output, not production use.

The example was checked with Timoni 0.35.0, which embeds CUE 0.17.1. It
needs no separate CUE executable, registry or Kubernetes cluster.

The [Kustomize comparison](../example-kustomize/README.md) renders the same
Deployment with 2 or 3 replicas, and shows why rendering 0 differs from
violating this module's explicitly declared range.
