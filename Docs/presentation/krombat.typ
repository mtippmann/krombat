// Slides: typst compile --font-path Docs/presentation/fonts Docs/presentation/krombat.typ Docs/presentation/krombat.pdf
// Notes: typst compile --font-path Docs/presentation/fonts --input notes=true Docs/presentation/krombat.typ Docs/presentation/speaker-notes.pdf
#import "@preview/diatypst:0.9.3": slides as diatypst-slides

#let ink = rgb("14263d")
#let muted = rgb("52647a")
// Keep diatypst's blue, with pale blue cards for the content helpers.
#let accent = blue.darken(50%)
#let pale = accent.lighten(90%)
#let notes-mode = sys.inputs.at("notes", default: "false") == "true"
#let pair(a, b) = grid(columns: (1fr, 1fr), gutter: 14pt, a, b)
#let panel(title, body) = [
  #text(size: 11pt, weight: "bold", fill: accent)[#title]
  #v(5pt)
  #body
]
#let callout(body) = block(width: 100%, fill: pale, radius: 4pt, inset: 10pt, body)
#let source(url, label) = link(url)[#label]

#let deck = (
  (
    seconds: 40,
    title: "Timoni + CUE, alongside Flux",
    body: [
      #text(size: 21pt, weight: "bold", fill: accent)[Author → render → reconcile]
      #v(13pt)
      #table(
        columns: (1.15fr, 2.85fr), inset: 7pt,
        table.header([*Tool*], [*Responsibility*]),
        [CUE + Timoni], [Define constrained inputs; build and package objects.],
        [Kustomize], [Transform YAML using bases, patches and replacements.],
        [Flux], [Continuously reconcile desired state in Kubernetes.],
      )
      #v(12pt)
      What changes when configuration has an explicit *contract*?
    ],
    refs: (
      source("https://timoni.sh/concepts", "Timoni"),
      source("https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/", "Kustomize"),
      source("https://fluxcd.io/flux/components/kustomize/kustomizations/", "Flux"),
    ),
    notes: [
      Start with the tools' responsibilities. CUE describes data and constraints;
      Timoni uses it to build and package Kubernetes objects. Kustomize transforms
      existing objects through bases, patches and other operations. Flux
      continuously reconciles declared state in the cluster. These responsibilities
      can be combined in a delivery pipeline. Our question is what an explicit
      configuration contract adds. We will use one replica count to explain
      constraints and composition, then apply the idea to a small kro API.
    ],
  ),
  (
    seconds: 40,
    title: "CUE’s creator: Marcel van Lohuizen",
    body: [
      - Member of the founding *Borg* team at Google.
      - Created *BCL*, the Borg Configuration Language, later generalized into GCL.
      - Created *CUE*, drawing on experience with large configurations.
      #v(18pt)
      #callout[
        #align(center)[#text(size: 20pt, weight: "bold", fill: accent)[Borg / BCL → GCL → CUE]]
      ]
      #v(12pt)
      The design idea: combine *data and constraints* in one language.
    ],
    refs: (
      source("https://cfp.cfgmgmtcamp.org/2024/talk/G9FQRT/", "Marcel’s biography"),
      source("https://cuelang.org/docs/introduction/", "CUE history"),
    ),
    notes: [
      Marcel van Lohuizen created CUE. At Google he was on the founding Borg
      team, where he created the Borg Configuration Language and core tooling.
      BCL was later generalized into GCL. CUE is a different language informed
      by that experience. Its design returns to constraint composition and graph
      unification. The useful connection for us: configuration can carry allowed
      values and reusable structure in the same language. The next two slides
      explain that model with small examples.
    ],
  ),
  (
    seconds: 80,
    title: "Types and values share one model",
    body: [
      Think of a constraint as a *set of allowed values*.
      #v(8pt)
      ```cue
      replicas: *2 | (int & >=1 & <=10)
      ```
      #v(9pt)
      - `int & >=1 & <=10`: integers in the intersection of these rules.
      - `|` offers alternatives; `*2` marks the preferred default.
      #v(13pt)
      #grid(
        columns: (1fr, 1fr, 1fr), gutter: 10pt,
        callout([No input → *2*]),
        callout([Input `3` → *3*]),
        callout([Input `0` → *error*]),
      )
    ],
    refs: (
      source("https://cuelang.org/docs/tour/basics/constraints/", "Constraints"),
      source("https://cuelang.org/docs/tour/types/defaults/", "Defaults"),
    ),
    notes: [
      Think of an integer type as the set of all integers, a range as a smaller
      set, and a concrete value as a singleton. CUE represents them within the
      same model. Read the expression from inside the parentheses first. The
      value must be an integer, at least one and at most ten. Ampersand means
      that every constraint must hold. The vertical bar offers alternatives,
      and the star marks two as the preferred default.

      Walk through the three cases. With no supplied replica value, evaluation
      selects two. Three satisfies the range and becomes the concrete result.
      Zero satisfies neither alternative, so evaluation fails. A string such
      as "three" also fails. The default is a preference, so it does not prevent
      valid input from selecting three. This expression comes from the included
      Deployment module. Exporting resources requires concrete values; a range
      alone does not choose a replica count. Pause on the cases before moving on.
    ],
  ),
  (
    seconds: 70,
    title: "Composition means satisfying both",
    body: [
      #pair(
        panel("Reusable definition + concrete input", [
          ```cue
          #Config: {
            replicas: int & >=1 & <=10
          }

          a: #Config & {replicas: 3}
          b: {replicas: 3} & #Config
          ```
        ]),
        panel("Unification", [
          Both `a` and `b` resolve to:
          ```cue
          replicas: 3
          ```
          #v(8pt)
          `A & B = B & A`
          #v(8pt)
          Conflicting information produces an error.
        ]),
      )
      #v(12pt)
      `#Config` names reusable structure and constraints.
    ],
    refs: (
      source("https://cuelang.org/docs/tour/basics/unification/", "Unification"),
      source("https://cuelang.org/docs/tour/basics/definitions/", "Definitions"),
      source("https://cuelang.org/docs/introduction/", "Composition properties"),
    ),
    notes: [
      The hash introduces a reusable CUE definition. Here Config contains a
      replica constraint without a default, so supplying a concrete value makes
      the narrowing operation clear. Combining Config with replicas three gives
      the structure shown on the right. Reversing those two operands gives the
      same result. CUE calls this operation unification.

      The computer-science connection is intersection: the result must satisfy
      both sources of information. Unification is commutative, associative and
      idempotent. Adding the same compatible constraint twice does not change
      the answer. Supplying zero conflicts with this range; combining replicas
      three with replicas four conflicts too. That is different from an override
      operation in which one concrete value replaces another. Defaults support
      valid choices, as on the previous slide. These algebraic properties describe
      CUE unification; they do not imply that patches, command-line processing or
      an entire delivery pipeline can run in arbitrary order.
    ],
  ),
  (
    seconds: 70,
    title: "Transform YAML or constrain inputs",
    body: [
      #pair(
        panel("Kustomize · kustomization.yaml excerpt", [
          ```yaml
          resources:
          - ../base
          replicas:
          - name: demo
            count: 3
          ```
        ]),
        panel("CUE · definition + input", [
          ```cue
          #Config: {
            replicas: *2 | (int & >=1 & <=10)
          }
          values: #Config & {replicas: 3}
          ```
        ]),
      )
      #v(8pt)
      #table(
        columns: (0.7fr, 1.4fr, 1.6fr), inset: 7pt,
        table.header([*Input*], [*Kustomize result*], [*CUE module result*]),
        [`3`], [Render `3`], [Accept `3`],
        [`0`], [Render `0`], [Reject: outside `1–10`],
      )
      #v(9pt)
      Kustomize can be paired with separate schema and policy checks.
    ],
    refs: (
      source("https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/", "Kustomize operations"),
      source("example-kustomize/README.md", "Offline comparison"),
    ),
    notes: [
      Both examples select three replicas. Kustomize starts from a Deployment
      with two replicas and transforms that field to three. The replicas
      transformer is an alternative to a patch for this field.
      On the right, CUE combines an input with a definition that includes a
      default and a permitted range. The Timoni module connects values.replicas
      to the Deployment's replica count.

      Now change the input to zero. Kustomize can render zero, which is a valid
      Kubernetes replica count. It does not encode our example's minimum of one.
      CUE rejects zero because that minimum is part of the module's contract.
      This shows where the rule is expressed: schema tools, policies and CI
      checks can also accompany Kustomize. Both approaches reuse configuration.
      CUE makes reusable
      data and constraints part of one language; Kustomize offers familiar
      object transformations. Use the matching offline examples to inspect both.
    ],
  ),
  (
    seconds: 70,
    title: "Timoni packages CUE modules",
    body: [
      - *Interface:* values, defaults and constraints in a module.
      - *Schemas:* vendor Kubernetes API and CRD definitions for validation.
      - *Distribution:* versioned CUE source modules in OCI registries.
      #v(11pt)
      #panel("Build and validate the local WebApp module", [
        ```sh
        timoni build rgd-demo ./example-rgd -n demo
        timoni mod vet ./example-rgd --name rgd-demo -n demo
        ```
      ])
      #v(11pt)
      Build produces YAML. Timoni also offers CLI deployment and inventory.
    ],
    refs: (
      source("https://timoni.sh/concepts", "Timoni concepts"),
      source("https://timoni.sh/module", "Modules"),
      source("https://timoni.sh/cue/module/custom-resources", "CRD schemas"),
    ),
    notes: [
      Timoni provides a Kubernetes workflow around CUE. A module exposes a
      configuration interface and connects those inputs to resources. Authors
      can vendor Kubernetes API or custom-resource schemas so the resource
      definitions are checked as well as their own configuration rules. This
      is explicit schema integration: a CUE file alone does not automatically
      know every Kubernetes API.

      The first command renders our local WebApp RGD module. The second runs
      module validation against its included schema. These local commands do
      not contact Kubernetes. Timoni can package CUE sources and included schemas
      as a versioned OCI artifact; consumers build it with their own values.
      Its CLI also offers apply and instance inventory. Continuous reconciliation
      remains a separate responsibility. The small Deployment example omits
      vendored Kubernetes schemas; the RGD example includes the upstream kro
      schema. Structural validation is useful evidence, but controller behavior
      still needs runtime testing.
    ],
  ),
  (
    seconds: 80,
    title: "Use shared CUE structure for a kro API",
    body: [
      #pair(
        panel("CUE · schema + shared conventions", [
          ```cue
          #RGD:
            rgdv1.#ResourceGraphDefinition & {
              spec: schema: {
                group: "platform.example"
                scope: "Namespaced"
              }
            }
          #WebAppRGD: #RGD & {
            metadata: name: "web-app"
            // API + resource templates
          }
          ```
        ]),
        panel("Rendered RGD · schema/template excerpt", [
          ```yaml
          schema:
            kind: WebApp
            spec:
              replicas: >-
                integer | default=2
                minimum=1 maximum=10
          resources:
          - id: webDeployment
            template:
              kind: Deployment
              spec:
                replicas: >-
                  ${schema.spec.replicas}
          ```
        ]),
      )
      #v(9pt)
      Timoni sets the API default *2*; kro uses each instance's value.
    ],
    refs: (
      source("https://kro.run/docs/concepts/rgd/overview/", "kro v0.9.4"),
      source("example-rgd/timoni.cue", "Complete module, including vendored schema"),
      source("example-runtime/README.md", "Flux values handoff"),
    ),
    notes: [
      kro turns a ResourceGraphDefinition into a Kubernetes API and reconciles
      each instance's resources. This example is a namespaced WebApp backed by
      one Deployment. CUE combines shared conventions and the vendored kro 0.9.4
      schema, then adds graph-specific fields. Both panels are excerpts; the
      full module includes the remaining Deployment fields.

      Read the right panel from schema to resource. Default two and the range
      are part of the generated API. A WebApp can supply replicas three; the CEL
      expression reads three and kro is expected to reconcile three replicas.
      Changing the API default does not rewrite existing instances.

      Timoni preserves the expression as data; kro evaluates it later. The RGD
      is cluster-scoped; WebApps and their Deployments are namespaced. Runtime
      behavior remains illustrative. When delivering RGDs through Flux, preserve
      CEL expressions during Flux substitution; the handoff example explains how.
    ],
  ),
  (
    seconds: 90,
    title: "Where Bicep values meet Kubernetes",
    body: [
      #panel("Existing infrastructure handoff", [
        #callout[#align(center)[Bicep / CI → ConfigMaps / Secrets → Flux → Kubernetes]]
      ])
      #v(12pt)
      #pair(
        panel("Flux Kustomization · excerpt", [
          ```yaml
          postBuild:
            substituteFrom:
            - kind: ConfigMap
              name: infra-values
            - kind: Secret
              name: infra-secrets
          ```
        ]),
        panel("Where CUE can help", [
          - Validate outputs before publishing values.
          - Or: Timoni Runtime reads cluster values into typed Bundle inputs.
        ]),
      )
      #v(10pt)
      #text(size: 14pt)[CUE must receive the values that Flux will substitute.]
    ],
    refs: (
      source("https://fluxcd.io/flux/components/kustomize/kustomizations/#post-build-variable-substitution", "Flux substitution"),
      source("https://timoni.sh/bundle-runtime", "Timoni Runtime"),
      source("example-runtime/README.md", "Offline handoff example"),
    ),
    notes: [
      This is the infrastructure handoff described by the team: Bicep or CI
      creates ConfigMaps and Secrets, and Flux reads those values. The object
      names shown are illustrative. Flux's substituteFrom loads values for
      substitution after Kustomize has built the manifests, during reconciliation.

      The important question is when values become concrete. CUE can check
      required fields, types and relationships once it receives actual values.
      Validating a placeholder in CI does not validate the value Flux inserts
      later. Bicep also has typed outputs; the additional contract belongs at
      the boundary to the Kubernetes consumer.

      There are two possible additions. Keep Flux and validate producer outputs
      before publishing them, or validate the resolved consumer input. Alternatively,
      Timoni Runtime can read existing cluster objects into typed inputs for a
      Bundle, which groups module instances. That is a CLI workflow, not something
      Flux automatically invokes. An added render step also needs a trigger when
      its input values change. The offline example checks synthetic runtime
      inputs and rejects invalid or missing values. It does not fetch real objects
      or demonstrate production integration. Trace the handoff before choosing
      where validation belongs.
    ],
  ),
  (
    seconds: 60,
    title: "Choose the abstraction you need",
    body: [
      #pair(
        panel("What CUE + Timoni offer", [
          - Explicit, reusable configuration contracts.
          - Constraint composition and errors before deployment.
          - Modules with schema checks and OCI distribution.
        ]),
        panel("What the team takes on", [
          - CUE's model and error diagnostics.
          - Module, schema and toolchain maintenance.
          - Validation at the external-values handoff.
        ]),
      )
      #v(14pt)
      #callout[Use Kustomize for straightforward transformations. Consider CUE when shared rules need to compose.]
      #v(12pt)
      #text(size: 11pt)[
        Try it: #link("example/README.md")[Deployment] ·
        #link("example-kustomize/README.md")[Kustomize comparison] ·
        #link("example-rgd/README.md")[WebApp RGD]
      ]
    ],
    refs: (),
    notes: [
      Close on the tradeoff. CUE lets configuration authors express reusable
      structure and contracts together. Unification combines compatible rules
      and exposes contradictions during evaluation. Timoni turns that language
      into modules, schema integration and a Kubernetes packaging workflow.
      These benefits come with a language to learn, diagnostics to understand,
      and schemas and toolchains to maintain.

      Kustomize remains a good fit for straightforward transformations, especially
      when existing validation already covers the required policies. Flux can
      continue reading infrastructure values and reconciling resources. CUE is
      most interesting when several sources of shared structure and rules must
      compose reliably. The README links offline output comparisons, the
      runtime-input contract and additional study material. Finish at ten
      minutes and invite questions afterward.
    ],
  ),
)

#let total-seconds = deck.map(slide => slide.seconds).sum()
#assert(deck.len() == 9, message: "The talk must contain nine slides.")
#assert(total-seconds == 600, message: "The speaking plan must total 10:00.")
#let stamp(s) = str(calc.floor(s / 60)) + ":" + (if calc.rem(s, 60) < 10 { "0" } else { "" }) + str(calc.rem(s, 60))

#set text(font: "Fira Sans", fill: ink)
#set par(leading: 0.4em, spacing: 0.6em)

#if notes-mode {
  set document(title: "Timoni and CUE — speaker notes")
  set page(
    paper: "a4", margin: 22mm,
    footer: context align(right)[#text(size: 9pt, fill: muted)[Speaker notes · #counter(page).display() / #deck.len()]],
  )
  set text(size: 12pt)
  show raw: set text(font: "Fira Code", size: 10pt)
  let start = 0
  for (i, slide) in deck.enumerate() {
    if i > 0 { pagebreak() }
    text(size: 10pt, fill: accent, weight: "bold")[Slide #(i + 1) · #stamp(start)–#stamp(start + slide.seconds)]
    heading(level: 1, slide.title)
    slide.notes
    v(12pt)
    text(size: 10pt, fill: muted)[Timing: #slide.seconds seconds. Planned total: #stamp(total-seconds).]
    if slide.refs.len() > 0 {
      v(14pt)
      text(size: 10pt)[Sources: #slide.refs.join([ · ])]
    }
    start += slide.seconds
  }
  context {
    assert(counter(page).final().first() == deck.len(), message: "Speaker notes must occupy exactly nine pages.")
  }
} else {
  show: diatypst-slides.with(
    title: "Timoni and CUE",
    subtitle: "Constraints, composition and Kubernetes",
    layout: "large",
    ratio: 16/9,
    theme: "normal",
    count: "number",
    footer-title: "Timoni + CUE",
    footer-subtitle: "Constraints · composition · Kubernetes",
    first-slide: false,
    toc: false,
  )
  set text(size: 16pt)
  set list(indent: 13pt, body-indent: 5pt, spacing: 8pt)
  show raw: set text(font: "Fira Code", size: 11pt)
  for slide in deck {
    heading(level: 2, slide.title)
    slide.body
    if slide.refs.len() > 0 {
      v(1fr)
      v(4pt)
      text(size: 7.5pt, fill: muted)[Sources: #slide.refs.join([ · ])]
    }
  }
  context {
    assert(counter(page).final().first() == deck.len(), message: "Slide overflow: expected exactly nine PDF pages.")
  }
}
