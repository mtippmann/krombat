# Project review: evidence and boundaries

Review date: 2026-10-27. Baseline: `7a932123160cac65f9cae9679866eba897c8a642`
(`add timoni comparision`). The checkout was clean before this work.

## Scope

Static review covered application architecture and request/state flow,
resource-inspector ownership checks, manifests and RBAC, local deployment
helpers, test entry points, CI, and the README/demo/runbook/workshop/Timoni
documentation. Detailed regression review focused on the latest commit's
handwritten Timoni templates, vendored schema integration, tests, packaging,
and newly added documentation. Generated Kubernetes schemas were checked
as dependencies, not manually audited field by field.

No application, manifest, deployment, or test code was changed. The only
repository additions are this presentation directory. No cluster was
contacted, no production service was contacted, no controller was installed
or changed, and no artifact was published.

## Findings against the Timoni change

**No confirmed introduced correctness defect was found in the reviewed
change.** This is not a declaration that every existing project behavior
is bug-free, nor a live-deployment approval. Existing baseline defects are
outside the introduced-regression finding set.

The following properties are supported by source inspection and offline checks:

| Area | Evidence / conclusion |
|---|---|
| Runtime boundary | The module emits six RGDs, not game-instance snapshots. CEL and graph conditions remain runtime inputs to kro. |
| YAML compatibility | All six default graph objects match the original YAML structurally; only the upstream implicit namespaced scope is made explicit. |
| CUE factoring | Shared envelopes and resource templates preserve graph-specific IDs, expressions and resource order in the baseline comparison. |
| Configuration | Override tests restrict output changes to labels and Dungeon API defaults; representative invalid values fail. |
| Application boundary | Go owns root turn calculations and saves `spec.game`; the backend RBAC does not grant write access to projected child resources. |
| Release ownership | The guides distinguish cluster-scoped graph ownership from application and gameplay lifecycles, and explicitly prohibit prototype adoption over existing graphs. |
| OCI / Flux explanation | Documentation distinguishes the reusable Timoni source module from rendered manifests delivered by Flux. Flux/application packaging is described as future work, not implemented behavior. |
| Documentation navigation | Local Markdown link targets exist in all 16 checked README/Docs/module README files. External URLs and section anchors were not automatically validated. |

## Commands actually run

| Check | Result |
|---|---|
| `bash timoni/test.sh` | Pass: format; default/debug vet; four Python test methods covering graph parity, override behavior, representative rejection cases and local OCI packaging. |
| `bash tests/guardrails.sh` | Pass. |
| `bash tests/local-script.sh` | Pass with stubbed tools; no live cluster contacted. |
| Local Markdown file-link scan | Pass: no missing local target files in the 16 checked Markdown files. |
| `docker info --format '{{.ServerVersion}}'` | Blocked: permission denied accessing `/var/run/docker.sock`. |
| Typst slide and note compilation | Pass with Typst 0.15.1; 16 slide pages and 16 note pages. |

Installed Timoni: **0.35.0**, embedded CUE **0.17.1**. The guides specify
Timoni **0.34.0**. Results here establish behavior on the installed client;
they do not independently reproduce the exact documented version.

The invalid-label-key probe also failed as expected at Timoni object
extraction: the reusable CUE label constraint is not the only validation
layer. No label-validation defect is inferred from that constraint alone.

## What remains unverified

- Docker application builds, backend unit tests/`go vet`, and frontend compilation.
- `tests/run-all.sh`, API integration, browser journeys, and live cleanup.
- Actual kro activation of the rendered prototype and any live Timoni inventory adoption.
- Upgrade readiness gates, backwards-compatible schema evolution, rollback and data recovery.
- Windows/Docker Desktop execution.
- External documentation URL availability and the exact pinned Timoni CLI version.

The full test entry point was not run because Docker is inaccessible and
its later steps also contact a deployed cluster. Do not interpret the
stubbed script checks as proof that a local cluster is ready.

## Follow-up validation, not newly discovered defects

1. On a machine with Docker and an approved local cluster, run the prescribed
   build, deploy and full test helpers before presenting the live demo.
2. Reproduce the offline module checks with the documented Timoni client,
   or deliberately update the supported version after testing it.
3. Consider including the standalone Timoni checks in CI with pinned tooling
   and Python dependencies; the existing container-check workflow does not
   currently exercise the new module.
4. Before any future live delivery pilot, define a single graph owner and
   explicit compatibility, health and recovery criteria. Test in a separately
   approved disposable local environment rather than adopting shared graphs.

## Presentation

The accompanying [Typst source](krombat.typ), [slides](krombat.pdf), and
[speaker notes](speaker-notes.pdf) turn the verified architecture and Timoni
walkthrough into a 20-minute talk. They distinguish implemented behavior,
review evidence and proposed future lifecycle features.

## Presentation revision: operations and platform audience

Reworked on 2026-10-27 at the user's request. The revised 16-slide talk
assumes Kubernetes, operators, Helm and YAML knowledge, then introduces
CUE, Timoni and kro. The focus is infrastructure definition maintenance,
OCI/ACR distribution, version/digest pinning, promotion, lifecycle and
ownership. The game is now only the final 45-second self-study handoff.
Speaker notes and the running order were rewritten; the total remains
20:00 with a short question slot.

Technical capabilities were checked against the official sources listed in
[SOURCES.md](SOURCES.md), including Microsoft's confirmation of ACR OCI
artifact support. ACR compatibility is a proposed use of those capabilities,
not a tested registry integration. The presentation explicitly acknowledges
Helm's OCI support and optional values schemas. Registry publication,
verification, Flux and live Timoni delivery remain proposed workflows.

Revision validation:

- `bash timoni/test.sh`: pass with installed Timoni 0.35.0 / CUE 0.17.1;
  default/debug vet and all four test methods pass.
- Both Typst outputs rebuilt with 0.15.1: 16 slide pages and 16 note pages.
- Slide timings sum to 1,200 seconds. A rendered contact sheet was inspected
  for readability and overflow; each slide occupies one PDF page.
- No application code or deployment configuration changed. No Docker build,
  cluster command, registry login/push, controller change or remote publish
  was performed for this presentation revision.

The original review above is retained as historical evidence. It is not a
claim that its application checks were rerun during this revision.

## Presentation revision: paced introduction

The previous presentation framing was superseded at the user's request.
The current 20-minute deck introduces Timoni and CUE before using either
in a platform example. It presents CUE data, constraints, defaults,
unification and definitions in sequence, then follows one Deployment
through a Timoni module, values and generated YAML. Helm comparison,
OCI/ACR, digest pinning and kro come afterward. The game is mentioned only
in the last 30 seconds as optional self-study.

A small self-contained [teaching module](example/README.md) was added in
this directory. With Timoni 0.35.0 / CUE 0.17.1 and no cluster, its default
build emitted two replicas, `three.yaml` emitted three, `invalid.yaml` was
rejected, and `timoni mod vet` passed. The example has no vendored
Kubernetes schemas; it teaches configuration and rendering rather than
claiming full schema or live-cluster validation.

Both slide and notes PDFs were rebuilt with Typst 0.15.1. The 16 slide
timings total 20:00; slides and notes each occupy 16 pages. Slides were
inspected for readability and overflow. The registry example is documented
as a future workflow and was not published, signed or deployed. Original
review findings above remain historical evidence, not new test results.

## Presentation revision: expanded kro example

Added three slides at the end of the tool introduction. They show a WebApp
RGD's API and Deployment template, the CUE definition that combines a
vendored RGD schema with shared conventions, and an illustrative WebApp
instance and resulting Deployment. Earlier slide timings were adjusted;
the talk remains 20 minutes with 19 slides.

Added a separate [WebApp RGD teaching module](example-rgd/README.md), using
this checkout's vendored kro v0.9.4 RGD schema. The presentation identifies
its runtime behavior as illustrative. Its default and override builds and
CRD-schema vetting are offline checks; they do not demonstrate RGD activation
or reconciliation by kro. No registry or cluster was contacted.

## Current presentation revision: 10-minute technical comparison

Revised on 2026-10-29. This section supersedes earlier presentation
descriptions; the review findings and validation records above remain
historical evidence.

The current deck has nine slides for a computer-science-oriented platform
and developer audience. It introduces Marcel van Lohuizen and the BCL/GCL
background, explains constraints and unification, and compares CUE input
contracts with Kustomize transformations. Timoni's module and schema
workflow leads into the existing kro WebApp example. Flux is presented as
the continuous reconciler of rendered manifests in a proposed delivery
design. Business-domain framing and standalone Helm, ACR and digest
lessons were removed from the talk; registry material remains optional.

The source now uses diatypst 0.9.3, pinned explicitly, with a normal theme,
16:9 layout and numbered slides. It retains a separate A4 notes mode and
asserts nine records, 600 seconds and nine physical pages in each output.
The first package download requires network access; subsequent compilation
can use the cached package. No application or existing module behavior
was changed. A separate [Kustomize example](example-kustomize/README.md)
was added for the comparison.

Validation performed for this revision:

- Slides and notes compile with Typst 0.15.1 and diatypst 0.9.3. Each PDF
  has nine pages. All rendered slides and notes were inspected for fit and
  readability. Notes total 10:00, with approximately 1,240 narration words;
  a text-based pacing audit gives about 119–132 words per minute by slide.
- Timoni 0.35.0 / CUE 0.17.1 renders 2 and 3 replicas and rejects 0 for
  the Deployment module; `timoni mod vet` passes. kubectl 1.37.1 / embedded
  Kustomize 5.8.1 renders 2, 3 and 0. The 2- and 3-replica Deployments are
  structurally identical between tools. Every kubectl invocation uses an
  explicit `minikube` context and is client-only.
- The displayed CUE definition/input snippets resolve to 3 in temporary
  module adapters. Both unification operand orders produce 3; conflicting
  concrete values are rejected.
- The WebApp module renders API defaults of 2 and 3 while preserving
  `${schema.spec.replicas}`. Its vendored CRD-schema vet passes. The folded
  YAML schema/CEL excerpts match the rendered values. The shared CUE RGD
  excerpt composes with the full WebApp definition when its omitted API
  version is restored.
- Local file links in all seven presentation Markdown documents and in
  the Typst source resolve.

No application build, live cluster check, controller change, registry
publication, Flux integration or remote modification was performed. Runtime
results remain illustrative. These checks use the installed Timoni 0.35.0,
not the separate Krombat guide's documented 0.34.0.

## Follow-up: Bicep / Flux values handoff

Updated on 2026-10-29 after the user clarified that Bicep or CI creates
ConfigMaps/Secrets that Flux reads. Slide 8 now explains that handoff,
Flux's post-build substitution and the boundary at which CUE can validate
concrete inputs. The object names are illustrative; no company Bicep or
Flux configuration was inspected. Slide 7 retains the API-default/runtime
distinction, and the closing notes account for infrastructure inputs.
The deck remains nine slides and 600 seconds using diatypst 0.9.3.

The [runtime-input example](example-runtime/README.md) demonstrates Timoni's
native ConfigMap/Secret query mapping and typed Bundle inputs. It uses
`bundle vet --offline` with synthetic environment values and an explicit
`minikube` context. Runtime inputs are bound separately and unified with
the consuming CUE contract. The module URL is a placeholder; this check
does not fetch or validate a referenced module. Timoni Runtime remains a
CLI workflow; no automatic Flux invocation or change propagation is
claimed. The guide also explains preserving kro CEL through Flux
substitution.

Validation for this follow-up:

- Timoni 0.35.0 / CUE 0.17.1: the valid input passes; values below or above
  the range, fractional and nonnumeric values, missing number/token and
  an empty token are rejected. All eight cases use `--offline`.
- `timoni fmt --diff example-runtime` passes. The displayed Flux excerpt
  parses as YAML and references both ConfigMap and Secret inputs.
- Both PDFs compile and contain nine pages; timings remain 10:00.
  Changed slides and notes were inspected for fit and readability.
- Local file links in all eight Markdown documents and the Typst source
  resolve.

No live runtime lookup, cluster command, Azure operation, registry access,
secret retrieval or production integration was performed for this update.
Earlier module and application validation records remain historical.

## Follow-up: default serif typography and blue

Updated on 2026-10-29 at the user's request. Removed the custom text/raw
font families, base font sizes and paragraph/list typography overrides.
The deck and notes now inherit Typst's default serif text and code
styling. Diatypst supplies its default blue title color and white
background; content helpers derive their colors from the same blue.
The nine-slide structure, 600-second plan and presentation content are
unchanged.

Both PDFs compile with Typst 0.15.1 and diatypst 0.9.3, with nine pages
each. Embedded-font inspection confirms Libertinus Serif regular/bold;
the deck uses the default DejaVu Sans Mono for code. All slides and a
notes page were visually checked after the font change. README typography
details and the source list were updated. No application, cluster or
runtime-input checks were needed for this styling-only change.

## Current state: consistency fixes and Fira typography

Updated on 2026-10-29. The current deck restores larger sans-serif
typography with Fira Sans for body/headings and Fira Code for code.
It retains diatypst 0.9.3, dark blue headings, pale blue callouts, a white
background, nine slides and the 600-second speaking plan. Regular and
bold font files are bundled with their licenses; the documented build
commands use `--font-path Docs/presentation/fonts`.
PDFs and font assets are marked as binary in Git.

The environment-file comment now lists both supported local contexts.
The current Timoni guides and presentation agree on the verified
0.35.0 / CUE 0.17.1 toolchain, and the repository README links the deck.
Existing manual context examples were left unchanged at the user's
request. Earlier version and typography records above remain historical.

Validation for this update:

- `bash timoni/test.sh` passes: formatting, default/debug schema vetting
  and all four Python test methods, including graph parity and local OCI
  packaging. No cluster is contacted.
- Both PDFs compile with Typst 0.15.1 and contain nine pages. Font
  inspection confirms embedded Fira Sans regular/bold and Fira Code.
  Compilation also succeeds with system fonts disabled and the bundled
  fonts supplied. Slide text remains within page bounds; all slides and
  a notes page were visually checked.
- Speaker-note timings total 600 seconds. Local file links in all 27
  repository Markdown documents resolve. Current version references agree
  on 0.35.0, and the manual context commands match their previous contents.

No Docker, Kubernetes or Azure operations were performed. The font files
were obtained or copied for reproducible local builds; no artifact was
published and no remote repository was modified.
