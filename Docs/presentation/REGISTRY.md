# Optional: publishing and pinning a module

The 15-minute talk mentions OCI module distribution. This optional page
expands on publishing and digest pinning for the small
[presentation example](example/README.md).

These commands are **illustrative and were not executed**. Replace
`example.azurecr.io` with an authorized registry and authenticate using its
supported credential process.

From the repository root:

```bash
# Publish a reviewed module as an OCI artifact.
timoni mod push ./Docs/presentation/example \
  oci://example.azurecr.io/platform/web \
  --version 1.0.0 --latest=false

# Look up the published version's digest.
timoni mod list oci://example.azurecr.io/platform/web
```

The module reference belongs inside a bundle instance:

```cue
bundle: {
    apiVersion: "v1alpha1"
    name: "example-apps"
    instances: web: {
        module: {
            url: "oci://example.azurecr.io/platform/web"
            version: "1.0.0"
            digest: "sha256:<actual-digest>"
        }
        namespace: "demo"
        values: replicas: 3
    }
}
```

The URL and digest are placeholders. Replace the digest with the actual
published artifact digest before using the bundle. A version tag is a
human-readable reference; when both a version and digest are provided,
Timoni verifies the retrieved artifact matches the digest. To reproduce a
rendered release, also record values and toolchain; pin any workload images
separately. Verification of a trusted publisher requires a separate signing
and verification policy.

A module OCI artifact carries source and any bundled schemas. Rendered YAML
is a different output. In a [Flux delivery workflow](https://timoni.sh/gitops-flux),
CI can render YAML and publish a separate manifest artifact for Flux to
reconcile. Pick one owner for applying and pruning each Kubernetes object.

Publishing a module does not deploy it. This example does not authorize
publishing to the original remote or adopting existing shared resources.
See the repository's local-cluster rules before any future cluster operation.

Sources: [Timoni publishing](https://timoni.sh/cue/module/publishing),
[Bundle digest references](https://timoni.sh/bundle#digest),
[module signing](https://timoni.sh/cue/module/signing), and
[ACR content formats](https://learn.microsoft.com/en-us/azure/container-registry/container-registry-image-formats).
