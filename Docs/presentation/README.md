# Timoni and CUE: constraints, composition and Kubernetes

A **10-minute talk** for platform engineers and developers familiar with
Kubernetes, YAML and Flux/Kustomize. It assumes a computer-science
background, but no prior knowledge of CUE, Timoni or kro.

- [Slides (PDF)](krombat.pdf): nine widescreen slides using diatypst.
- [Speaker notes (PDF)](speaker-notes.pdf): timed, one page per slide.
- [Editable Typst source](krombat.typ).
- Offline examples: [Deployment](example/README.md),
  [Kustomize comparison](example-kustomize/README.md),
  [WebApp RGD](example-rgd/README.md) and
  [Bicep/Flux runtime-input handoff](example-runtime/README.md).
- [Technical sources](SOURCES.md), [optional registry material](REGISTRY.md)
  and [project review / revision evidence](REVIEW.md).

The talk introduces CUE's data-and-constraints model, connects unification
to intersection, and compares that with Kustomize's object transformations.
It then shows what Timoni adds and uses a kro WebApp API as a compact
application. It also examines the team's Bicep/CI → ConfigMaps/Secrets →
Flux flow and where concrete infrastructure values can be validated.

## Running order

| Slide | Time | Topic |
|---|---|---|
| 1 | 00:00–00:40 | CUE/Timoni, Kustomize and Flux responsibilities |
| 2 | 00:40–01:20 | Marcel van Lohuizen: Borg, BCL/GCL and CUE |
| 3 | 01:20–02:40 | Types, constraints, values and defaults |
| 4 | 02:40–03:50 | Reusable definitions and unification |
| 5 | 03:50–05:00 | Kustomize transformations versus CUE constraints |
| 6 | 05:00–06:10 | Timoni modules, schemas, build and packaging |
| 7 | 06:10–07:30 | Shared CUE structure for a kro API |
| 8 | 07:30–09:00 | Bicep values and a typed Timoni Bundle input contract |
| 9 | 09:00–10:00 | Benefits, costs and follow-up examples |

The speaking plan totals 600 seconds. Questions follow the talk.
The source asserts both the nine-slide count and the timing total.

## Presenter guidance

Explain one replica count throughout. Read `&` as satisfying both sources
of information, `|` as alternatives and `*` as a preferred default. The
set/intersection model is the technical anchor; no lattice-theory detour
is needed. Unification's order independence is a property of that CUE
operation, not a claim about every step in a deployment pipeline.

On slide 5, both tools produce the same Deployment for counts 2 and 3.
Kustomize can also render 0, a valid Kubernetes value. The CUE module
rejects it because it explicitly encodes the teaching example's 1–10
range. Explain where the rule lives; Kustomize can have separate schema,
policy and CI checks. Avoid claiming that validation or reuse is
exclusive to CUE.

On slide 7, distinguish the RGD release default from an individual
WebApp's runtime value. Timoni emits the RGD and preserves its CEL
expressions. kro evaluates those expressions for instances later. The
RGD is cluster-scoped; WebApp instances and their Deployments are
namespaced. The CUE and YAML panels are explicitly shortened excerpts.

On slide 8, follow the described Bicep/CI → ConfigMaps/Secrets → Flux
handoff. The object names are illustrative. Flux can resolve values after
rendering; earlier CUE checks of placeholders cannot validate those
concrete values.
Validation must receive concrete inputs at the producer or consumer
boundary. Timoni Runtime can read cluster objects into typed Bundle
inputs. The excerpt separates Runtime bindings (`_replicas`, `_token`)
from the consumer constraints: an integer from 1 to 10 and a nonempty
string. Bundle metadata and the module reference are omitted; the
complete example factors the same constraints into `#InfraValues`.
This is an explicitly invoked CLI workflow. Flux does not
automatically run it; an additional render step needs its own trigger
when the inputs change. The [handoff example](example-runtime/README.md)
demonstrates the runtime contract offline and explains preservation of
kro CEL expressions through Flux substitution.

No production integration is implemented here. API readiness must precede
WebApp instances; expected runtime results remain illustrative. A build
or structural schema check does not establish live controller behavior.

Use the prepared slides rather than running a live demo. Rehearse with
the notes, leaving time to point at the code and trace the diagrams.
The notes are speaking guidance, not text that must be read verbatim.

## Reproduce the examples offline

From this directory:

```bash
# CUE default and valid override: 2 and 3 replicas.
timoni build demo ./example -n demo
timoni build demo ./example -n demo --values ./example/three.yaml

# Expected failure: 0 violates the module's range.
timoni build demo ./example -n demo --values ./example/invalid.yaml

# Kustomize renders 2, 3 and 0; no cluster is contacted.
kubectl --context=minikube kustomize example-kustomize/base
kubectl --context=minikube kustomize example-kustomize/three
kubectl --context=minikube kustomize example-kustomize/zero

# RGD build and vendored CRD-schema validation.
timoni build rgd-demo ./example-rgd -n demo
timoni build rgd-demo ./example-rgd -n demo --values ./example-rgd/api-default-three.yaml
timoni mod vet ./example-rgd --name rgd-demo -n demo

# Native Runtime injection using dummy inputs; --offline prevents API reads.
DEMO_REPLICAS=3 DEMO_TOKEN=demo-only \
  timoni bundle vet -f example-runtime/bundle.cue -r example-runtime/runtime.cue \
  --offline --kube-context=minikube
```

All commands above are local rendering or validation. The explicit
`minikube` context follows the repository convention and does not change
the current context; `docker-desktop` is the other permitted choice.

Checked with Timoni 0.35.0 / embedded CUE 0.17.1 and kubectl 1.37.1 /
embedded Kustomize 5.8.1. The Krombat module guides use the same Timoni
0.35.0 / CUE 0.17.1 toolchain.
The Deployment module omits vendored Kubernetes schemas; the RGD module
includes the schema from unmodified upstream kro v0.9.4.

## Rebuild the PDFs

From the repository root:

```bash
typst compile --font-path Docs/presentation/fonts Docs/presentation/krombat.typ Docs/presentation/krombat.pdf
typst compile --font-path Docs/presentation/fonts --input notes=true Docs/presentation/krombat.typ Docs/presentation/speaker-notes.pdf
```

Fira Sans and Fira Code are bundled in [fonts](fonts/README.md), with their
licenses. The `--font-path` option makes these fonts available to Typst.

The source imports [diatypst 0.9.3](https://typst.app/universe/package/diatypst/0.9.3/).
The first compile requires access to Typst's package registry unless the
package is already cached. Later builds can use the cached package.
To select another cache location, pass `--package-cache-path <directory>`
to both commands. The package requires Typst 0.12.0 or later; this deck
was built with 0.15.1.

The deck uses diatypst's normal theme, dark blue headings, pale blue
callouts and a white background, with 16:9 pages and numbered slides.
There are no automatic title, contents or section pages. Fonts are
Fira Sans and Fira Code: body text is 16 pt, code 11 pt and
slide titles 22.4 pt on a 213 × 120 mm page. Speaker notes use 12 pt
body text. References use smaller text at the bottom.

## Further study

- [CUE language tour](https://cuelang.org/docs/tour/).
- [Kustomize operations](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/).
- [Flux Kustomizations](https://fluxcd.io/flux/components/kustomize/kustomizations/).
- [Timoni concepts](https://timoni.sh/concepts) and [module documentation](https://timoni.sh/module).
- [Bicep/Flux handoff and offline Timoni Runtime example](example-runtime/README.md).
- [Optional OCI/ACR and digest material](REGISTRY.md).
- [Krombat Timoni guide](../timoni/README.md), [kro workshop](../workshop/README.md)
  and [local game demo](../demo/DEMO.md).

The talk needs no running cluster. Any later local game exploration must
follow the repository's minikube/Docker Desktop, ownership and cleanup
rules.
