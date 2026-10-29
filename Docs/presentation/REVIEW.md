# Validation evidence

Updated: 2026-10-29.

## Presentation

The [slides](krombat.pdf) and [speaker notes](speaker-notes.pdf) provide
a 15-minute introduction to CUE, Timoni, Kustomize and Flux. The twelve
slides cover constraints, unification, module/schema integration, a kro
API, and external values passed through Runtime queries and a Bundle
contract. Examples use synthetic values and generic resource names.

The [Typst source](krombat.typ) pins diatypst 0.9.3 and asserts twelve
slides, 900 seconds and twelve physical pages in each PDF. Typography
uses bundled Fira Sans and Fira Code, dark blue headings and pale blue
callouts. Body text remains 16 pt and code 11 pt; notes use 12 pt text.

Current validation:

- Both PDFs compile with Typst 0.15.1 and contain twelve pages.
- The running order totals 15:00 and agrees with the source and notes.
- Rendered slides and the expanded Runtime notes were inspected for fit.
  Text stays within the slide page bounds.
- Local links resolve in all 27 repository Markdown files and the Typst
  source. Git whitespace checks pass.
- Existing shell command examples outside the presentation retain their
  previous contents. Documentation wording changes do not alter demo
  commands, application code or deployment configuration.

## Runtime input contracts

The [Runtime example](example-runtime/README.md) declares ConfigMap and
Secret field mappings, binds named inputs to hidden CUE fields and
unifies the consumed values with `#InfraValues`. The slide and guide
excerpts agree with the complete query definitions.

All eight cases pass with Timoni 0.35.0 / embedded CUE 0.17.1:

| Input | Expected outcome |
|---|---|
| Replicas `3`, nonempty token | Accept |
| Replicas `0` | Reject minimum |
| Replicas `11` | Reject maximum |
| Replicas `3.5` | Reject integer type |
| Replicas `three` | Reject number parsing |
| Missing replicas | Reject incomplete input |
| Missing token | Reject incomplete input |
| Empty token | Reject nonempty-string contract |

The cases run against both the complete files and the displayed Runtime
and Bundle excerpts with only omitted metadata restored. Each invocation
uses `bundle vet --offline`, synthetic environment inputs and an explicit
`minikube` context. No Kubernetes API or registry is contacted.

The placeholder module URL is deliberate. These checks validate Bundle
structure and its consumer contract; fetching a module and validating its
interface or resource schemas remain separate operations. The examples
define neither a continuously triggered pipeline nor a live Flux integration.

## Supporting example evidence

Previous offline checks remain applicable to the unchanged examples:

- The [Deployment module](example/README.md) renders 2 and 3 replicas and
  rejects 0 under its declared 1–10 range.
- The [Kustomize comparison](example-kustomize/README.md) renders 2, 3
  and 0. The 2- and 3-replica objects match the CUE module structurally.
  Zero is valid Kubernetes configuration; its rejection is an example
  consumer rule.
- The [WebApp RGD module](example-rgd/README.md) renders API defaults 2
  and 3 while preserving `${schema.spec.replicas}`. Vetting passes
  against its vendored upstream kro v0.9.4 schema.
- CUE composition checks resolve both operand orders to 3 and reject
  conflicting concrete values.

These are rendering and schema checks. They do not establish controller
activation or live reconciliation.

## Historical repository review

The 2026-10-27 review used baseline
`7a932123160cac65f9cae9679866eba897c8a642`. Source inspection covered
application/state flow, ownership boundaries, manifests, RBAC, local
scripts, tests and the Timoni graph module.

Recorded checks, including the 2026-10-29 module rerun:

| Check | Recorded result |
|---|---|
| `bash timoni/test.sh` | Formatting, default/debug vet, four Python test methods, graph parity, invalid values and local OCI packaging pass. |
| `bash tests/guardrails.sh` | Pass. |
| `bash tests/local-script.sh` | Pass with stubbed tools. |
| Docker daemon access | Unavailable during the original review. |

The six rendered RGDs match the original YAML structurally, with explicit
namespaced instance scope. Go computes turns and writes Dungeon
`spec.game`; kro projects child resources. Timoni packages graph
definitions rather than game-state snapshots.

These historical checks were not rerun for the documentation revision.
Docker builds, application tests, live cluster journeys, Windows/Docker
Desktop execution, graph activation, inventory adoption, upgrades and
recovery remain outside the current validation. No artifact was published
and no remote repository was modified.
