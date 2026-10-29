# Sources for the 10-minute Timoni and CUE talk

The current presentation uses official documentation for tool behavior
and a speaker-supplied biography for the creator overview. Relevant
sources are linked on slides and in the matching notes.

| Slide | Supporting sources |
|---|---|
| 1 | [Timoni concepts](https://timoni.sh/concepts), [Kustomize operations](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/) and [Flux Kustomizations](https://fluxcd.io/flux/components/kustomize/kustomizations/): authoring/rendering and continuous reconciliation responsibilities. |
| 2 | [Marcel van Lohuizen's Config Management Camp biography](https://cfp.cfgmgmtcamp.org/2024/talk/G9FQRT/): founding Borg team, BCL and GCL. [CUE introduction and history](https://cuelang.org/docs/introduction/): CUE's roots and constraint-based design. The arrow diagram represents the lineage of ideas, not language compatibility. |
| 3 | [CUE constraints](https://cuelang.org/docs/tour/basics/constraints/) and [defaults](https://cuelang.org/docs/tour/types/defaults/). The expression is taken from the [Deployment module](example/timoni.cue). |
| 4 | [CUE unification](https://cuelang.org/docs/tour/basics/unification/), [definitions](https://cuelang.org/docs/tour/basics/definitions/) and [composition properties](https://cuelang.org/docs/introduction/). Order independence refers to unification, not a complete deployment workflow. |
| 5 | [Kustomize transformations](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/) and the local [Kustomize comparison](example-kustomize/README.md). Counts 2 and 3 produce equivalent Deployments; 0 renders in Kustomize and fails the CUE module's declared range. The range is an example contract, not a Kubernetes requirement. |
| 6 | [Timoni concepts](https://timoni.sh/concepts), [module specification](https://timoni.sh/module) and [custom-resource schema integration](https://timoni.sh/cue/module/custom-resources). The [WebApp module](example-rgd/timoni.cue) vendors the upstream kro v0.9.4 RGD schema. |
| 7 | [kro v0.9.4 RGD overview](https://kro.run/docs/concepts/rgd/overview/), [resource templates and expressions](https://kro.run/docs/concepts/rgd/resource-definitions/resource-basics/) and the local [WebApp example](example-rgd/README.md). Both code panels are excerpts; the CUE panel composes a vendored CRD schema with shared conventions. |
| 8 | The user-described Bicep/CI → ConfigMaps/Secrets → Flux flow; object names are illustrative. [Bicep outputs](https://learn.microsoft.com/en-us/azure/azure-resource-manager/bicep/outputs), [Flux post-build substitution](https://fluxcd.io/flux/components/kustomize/kustomizations/#post-build-variable-substitution), [Timoni Runtime](https://timoni.sh/bundle-runtime), [the Timoni/Flux integration boundary](https://timoni.sh/gitops-flux) and the [offline runtime-input example](example-runtime/README.md). CUE checks need concrete inputs; Flux does not automatically run Timoni Runtime. |
| 9 | Synthesis of the preceding examples and documentation, with links to the local [Deployment](example/README.md), [Kustomize comparison](example-kustomize/README.md) and [WebApp](example-rgd/README.md) examples. |

## Presentation package

[diatypst 0.9.3](https://typst.app/universe/package/diatypst/0.9.3/) supplies
slide boundaries, headers, footers, counters and styling. The source pins
that version and uses its normal theme with dark blue headings, pale
blue callouts and a white background. Typography uses Fira Sans and
Fira Code at the presentation sizes documented in the README. Font files,
provenance and licenses are included in [fonts](fonts/README.md).
Speaker notes use a separate A4 layout from the same content and timing
records.

## Evidence boundaries

Checks are offline: local builds, schema vetting, PDF compilation and
render inspection. No live reconciliation, registry publication or Flux
integration was performed. The talk demonstrates tool behavior, not a
recommendation to replace an existing delivery system.

The Bicep/Flux flow comes from the user's description, not inspected
company configuration. Native Runtime contract checks use synthetic
environment values with `--offline`; they do not establish cluster reads,
module-schema validation or continuous change propagation. The handoff
guide also documents Flux escaping or exclusion for kro CEL expressions.

For optional material removed from the 10-minute deck, see
[OCI/ACR and digest examples](REGISTRY.md). ACR artifact support is
supported by [Microsoft's content-format documentation](https://learn.microsoft.com/en-us/azure/container-registry/container-registry-image-formats);
no ACR integration was tested. Historical project-review evidence remains
in [REVIEW.md](REVIEW.md).
