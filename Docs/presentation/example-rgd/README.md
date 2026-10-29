# WebApp RGD example

This example accompanies slides 6–7 of the Timoni and CUE talk. It
shows a small platform API using the Deployment from the earlier example.
It is a teaching module, not a deployed application.

A Timoni build emits one `ResourceGraphDefinition` (RGD). The RGD describes:

- A namespaced `platform.example/v1alpha1` `WebApp` API with a `replicas`
  field. The default is 2; allowed values are 1–10.
- A Deployment template whose replica count comes from the WebApp instance
  via kro's `${schema.spec.replicas}` CEL expression.

The module's [timoni.cue](timoni.cue) uses a reusable `#RGD` envelope for
shared metadata and API conventions, then adds the WebApp-specific schema
and template. It combines the envelope with a vendored CUE schema generated
from the upstream kro v0.9.4 RGD CRD. The vendored schema was copied from
this checkout's existing `timoni/krombat` module.

## Two different inputs

At **module build time**, the Timoni value `apiDefaultReplicas` configures
the API default. CUE validates it and writes it into the RGD as the
WebApp API's default. This value applies when a new WebApp omits the field.

At **runtime**, a WebApp instance can supply an explicit `spec.replicas`.
kro evaluates the CEL expression against that instance.
For the illustrated [webapp.yaml](webapp.yaml), kro is expected to reconcile
a Deployment named `demo` with three replicas. This runtime result has not
been verified on a cluster.

For example, a module default of 2 and an instance value of 3 are compatible:
the default is only used when the instance omits the field. Changing a
module default is an API-definition change, not a command to change the
replica count of all existing instances.

## Render it offline

From this directory, with the Timoni CLI installed:

```bash
# Emit an RGD whose WebApp API defaults to two replicas.
timoni build rgd-demo . --namespace demo

# Emit an RGD whose WebApp API defaults to three replicas.
timoni build rgd-demo . --namespace demo --values api-default-three.yaml

# Check the module and its RGD against the vendored CRD schema.
timoni mod vet . --name rgd-demo --namespace demo
```

Compare the two `schema.spec.replicas` values in the output. The runtime
expression `${schema.spec.replicas}` remains unchanged in both. The
`--namespace` argument supplies Timoni's instance namespace; the RGD itself
is cluster-scoped, while the WebApp API in its schema is namespaced.

The examples were checked with Timoni 0.35.0 / embedded CUE 0.17.1. Builds
and vet are offline. They check the module and RGD structure; they do not
prove that kro compiles the RGD or reconciles the expected Deployment.
No registry or cluster is required for these commands. A live installation,
upgrade and any ownership or API compatibility questions would need a
separate local validation following this repository's cluster rules.

For delivery through Flux with Bicep-derived values, see the
[handoff example](../example-runtime/README.md). It explains how to
preserve kro's CEL expressions during Flux substitution.
