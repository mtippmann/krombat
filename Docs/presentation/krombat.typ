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
    seconds: 45,
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
      Contracts describe allowed inputs before resources are rendered.
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
      can be combined in a delivery pipeline. An explicit configuration contract
      describes which inputs are allowed before resources are rendered. A single
      replica count illustrates constraints and composition, followed by a small
      kro API and external values supplied through Timoni Runtime.
    ],
  ),
  (
    seconds: 45,
    title: "CUE’s creator: Marcel van Lohuizen",
    body: [
      - Founding contributor to *Borg* at Google.
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
      Marcel van Lohuizen created CUE. At Google he was a founding contributor
      to Borg, where he created the Borg Configuration Language and core tooling.
      BCL was later generalized into GCL. CUE is a different language informed
      by that experience. Its design returns to constraint composition and graph
      unification. The technical connection: configuration can carry allowed
      values and reusable structure in the same language. The next two slides
      explain that model with small examples.
    ],
  ),
  (
    seconds: 90,
    title: "Types and values share one model",
    body: [
      A constraint describes a *set of allowed values*.
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
    seconds: 90,
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
    seconds: 90,
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
      Kubernetes replica count. It does not encode this example's minimum of one.
      CUE rejects zero because that minimum is part of the module's contract.
      This shows where the rule is expressed: schema tools, policies and CI
      checks can also accompany Kustomize. Both approaches reuse configuration.
      CUE makes reusable
      data and constraints part of one language; Kustomize offers familiar
      object transformations. Use the matching offline examples to inspect both.
    ],
  ),
  (
    seconds: 75,
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

      The first command renders the local WebApp RGD module. The second runs
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
    seconds: 90,
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
    seconds: 60,
    title: "External values and Flux substitution",
    body: [
      #panel("Flux delivery pattern", [
        #callout[#align(center)[Bicep / CI → ConfigMaps / Secrets → Flux → Kubernetes]]
      ])
      #v(10pt)
      #pair(
        panel("Flux Kustomization.spec · excerpt", [
          ```yaml
          postBuild:
            substituteFrom:
            - kind: ConfigMap
              name: infra-values
            - kind: Secret
              name: infra-secrets
          ```
        ]),
        panel("Consumer validation", [
          - Flux substitutes values after the Kustomize build.
          - CUE checks concrete input values against a contract.
          - Runtime can read cluster values into a Timoni Bundle.
        ]),
      )
      #v(8pt)
      #text(size: 14pt)[Validate producer outputs or resolve inputs before checking the consumer contract.]
    ],
    refs: (
      source("https://learn.microsoft.com/en-us/azure/azure-resource-manager/bicep/outputs", "Bicep outputs"),
      source("https://fluxcd.io/flux/components/kustomize/kustomizations/#post-build-variable-substitution", "Flux substitution"),
      source("example-runtime/README.md", "Values handoff"),
    ),
    notes: [
      Infrastructure outputs can enter Kubernetes as ConfigMaps and Secrets,
      published by Bicep or CI. Flux's substituteFrom reads those objects after
      Kustomize has built the manifests. This diagram illustrates one delivery
      pattern; the object names and values are synthetic.

      Bicep can type its outputs, while ConfigMap data is string-valued. The
      consumer still needs to establish the intended type and any additional
      rules. A replica count might need to be an integer from one to ten.

      CUE can validate producer outputs before publication, or check consumer
      inputs once they are resolved. Checking an earlier placeholder cannot
      validate a value inserted later by Flux. Timoni Runtime provides another
      input path: an explicit CLI step reads cluster objects and binds values
      into a Bundle. The next slides follow the lookup, binding and contract
      checks in that path. Flux's substitution and reconciliation remain a
      separate execution model.
    ],
  ),
  (
    seconds: 90,
    title: "Runtime selects resources and fields",
    body: [
      #panel("Runtime queries · excerpt from runtime.cue", [
        ```cue
        runtime: values: [{
          query: "k8s:v1:ConfigMap:demo:infra-values"
          for: "DEMO_REPLICAS": "obj.data.replicas"
        }, {
          query: "k8s:v1:Secret:demo:infra-secrets"
          for: "DEMO_TOKEN": "obj.data.token"
        }]
        ```
      ])
      #v(10pt)
      - `query`: API version, kind, namespace and object name.
      - `for`: named input → expression selecting a field.
      - Bundle attributes choose how each input is parsed.
      #v(10pt)
      #callout[#text(size: 14pt)[Object field → named Runtime input → typed Bundle value]]
    ],
    refs: (
      source("https://timoni.sh/bundle-runtime", "Runtime queries and mappings"),
      source("example-runtime/runtime.cue", "Complete runtime.cue"),
    ),
    notes: [
      A Runtime definition declares which resources to read and how their fields
      map to named inputs. Read the first query from left to right: Kubernetes
      API v1, ConfigMap, namespace demo, object infra-values. Its for mapping
      evaluates obj.data.replicas and exposes the result as DEMO_REPLICAS.
      The second query selects the token field from a Secret and names the input
      DEMO_TOKEN. These names connect the Runtime to the Bundle.

      The mapping expressions select concrete field values. A ConfigMap can hold
      the text three; selecting that field does not yet establish the consumer's
      integer range. The Bundle's attributes specify parsing, and CUE constraints
      establish the consumer contract on the following slide. A similar mapping
      can select other scalar fields from Kubernetes resources.

      The complete runtime.cue also supplies API metadata and explicitly selects
      the minikube context. Those fields are omitted from this query excerpt.
      Runtime reads happen when the CLI is invoked. A later change to a ConfigMap
      requires another execution to refresh the inputs. The offline example uses
      environment variables with the same names, allowing contract checks without
      Kubernetes API access.
    ],
  ),
  (
    seconds: 105,
    title: "Bundle inputs: bind, then constrain",
    body: [
      #panel("bundle.cue · contract, bindings and instance values (metadata omitted)", [
        ```cue
        #InfraValues: {
          replicas: int & >=1 & <=10
          token:    string & !=""
        }
        bundle: {
          _replicas: number @timoni(runtime:number:DEMO_REPLICAS)
          _token:    string @timoni(runtime:string:DEMO_TOKEN)
          instances: web: values: #InfraValues & {
            replicas: _replicas
            token:    _token
          }
        }
        ```
      ])
      #v(8pt)
      #callout[#text(size: 14pt)[Parse Runtime inputs → unify with the contract → supply module values]]
    ],
    refs: (
      source("https://timoni.sh/bundle", "Bundles"),
      source("https://timoni.sh/bundle-runtime", "Runtime bindings"),
      source("https://cuelang.org/docs/tour/references/hidden/", "Hidden fields"),
      source("example-runtime/bundle.cue", "Complete bundle.cue"),
    ),
    notes: [
      A Bundle declares module instances and their input values. This excerpt
      shows three distinct parts. First, `#InfraValues` is the reusable consumer
      contract: replicas must be an integer from one to ten, and token must be a
      nonempty string. The complete file also declares the Bundle's API version,
      name, module URL, module version and target namespace.

      Second, the Timoni attributes bind the named Runtime inputs. The number
      attribute parses `DEMO_REPLICAS` as a CUE number; the string attribute retains
      `DEMO_TOKEN` as text. The underscore fields hold those inputs separately from
      the values supplied to the module. They are hidden CUE fields and are not
      exported as application configuration.

      Third, `values` unifies `#InfraValues` with the concrete bindings. Text three
      from the ConfigMap becomes number three, then satisfies the integer range.
      Zero parses as a number but conflicts with the minimum. A fraction parses
      as a number but fails the integer requirement. Nonnumeric text fails at
      parsing, and an empty token conflicts with the nonempty-string constraint.

      The distinction between parsing and the consumer contract matters. Keep
      the constraints on the consumed values, separate from Runtime injection.
      The same named contract can be reused with other input sources. Building
      the referenced module then checks its own interface and resource schemas;
      this local Bundle check does not fetch that placeholder module.
    ],
  ),
  (
    seconds: 75,
    title: "Check the contract offline",
    body: [
      #panel("From example-runtime/ · synthetic inputs", [
        ```sh
        DEMO_REPLICAS=3 DEMO_TOKEN=demo-only \
          timoni bundle vet -f bundle.cue -r runtime.cue \
          --offline --kube-context=minikube
        ```
      ])
      #v(8pt)
      #table(
        columns: (1.4fr, 1.6fr), inset: 5pt,
        table.header([*Input*], [*Contract result*]),
        [Replicas `"3"`, nonempty token], [Accept integer `3` and token],
        [Replicas `0`, `11`, `3.5` or `three`], [Reject range, type or parsing],
        [Empty or missing token], [Reject required string],
      )
      #v(8pt)
      #text(size: 14pt)[Runtime reads on invocation. Input changes need a new run; Flux reconciles separately.]
    ],
    refs: (
      source("https://timoni.sh/bundle-runtime", "Offline Runtime inputs"),
      source("https://timoni.sh/gitops-flux", "Timoni and Flux"),
      source("example-runtime/README.md", "Reproduce the checks"),
    ),
    notes: [
      Offline mode supplies Runtime inputs from environment variables instead
      of querying Kubernetes. The variable names match the Bundle attributes.
      With replicas three and a nonempty dummy token, bundle vet succeeds.
      The same contract rejects zero, eleven, fractions and nonnumeric text.
      Missing replicas, a missing token and an empty token also fail.

      These failures happen at different layers: nonnumeric text cannot be
      parsed as a number, a fraction conflicts with int, and zero conflicts with
      the minimum. This separation makes the input boundary explicit.

      Bundle vet checks this Bundle and its concrete contract. The module URL
      in the example is a placeholder; module builds and resource-schema checks
      are separate. The command is an offline validation step and applies nothing.

      Runtime is executed by the CLI, so changed source values need a new run.
      Flux's documented templating workflow delivers rendered manifests and
      does not automatically invoke Runtime injection. A pipeline can resolve,
      validate and render inputs before handing manifests to Flux, provided it
      defines when that pipeline runs. The prepared checks demonstrate input
      validation; they do not implement that delivery pipeline.
    ],
  ),
  (
    seconds: 45,
    title: "Capabilities and operating requirements",
    body: [
      #pair(
        panel("Capabilities", [
          - Reusable CUE contracts and composition.
          - Module builds and vendored schema checks.
          - OCI modules and Runtime value binding.
        ]),
        panel("Operating requirements", [
          - CUE, schema and toolchain maintenance.
          - Concrete inputs at the validation boundary.
          - A trigger for each Runtime execution.
        ]),
      )
      #v(14pt)
      #callout[Timoni builds and applies when invoked; Flux continuously reconciles.]
      #v(12pt)
      #text(size: 11pt)[
        Examples: #link("example/README.md")[Deployment] ·
        #link("example-kustomize/README.md")[Kustomize comparison] ·
        #link("example-rgd/README.md")[WebApp RGD] ·
        #link("example-runtime/README.md")[Runtime inputs]
      ]
    ],
    refs: (),
    notes: [
      CUE combines reusable structure with constraints. Timoni adds modules,
      schema integration, OCI packaging and explicit CLI execution. Runtime
      binds external values before module evaluation. These capabilities require
      maintained interfaces, schemas and an execution trigger when inputs change.

      Kustomize provides object transformations; Flux continuously reconciles
      declared state. Their validation and execution responsibilities can be
      combined with CUE and Timoni in a delivery workflow. The linked examples
      cover rendering, composition, a kro API and Runtime contracts. Finish at
      fifteen minutes; questions follow the talk.
    ],
  ),
)

#let total-seconds = deck.map(slide => slide.seconds).sum()
#assert(deck.len() == 12, message: "The talk must contain twelve slides.")
#assert(total-seconds == 900, message: "The speaking plan must total 15:00.")
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
    assert(counter(page).final().first() == deck.len(), message: "Speaker notes must occupy exactly twelve pages.")
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
    assert(counter(page).final().first() == deck.len(), message: "Slide overflow: expected exactly twelve PDF pages.")
  }
}
