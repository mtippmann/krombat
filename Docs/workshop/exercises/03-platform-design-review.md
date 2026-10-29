# 03 - Platform design review

**Goal:** Compare kro resource-composition patterns, API boundaries and
operating requirements using the local resource graph.
This is a design exercise; it does not ask you to modify code or deploy
resources.

Consider a Kubernetes API that composes an application and its supporting
resources. Use the current game to examine composition, constraints, status
and reconciliation. Its local implementation provides the reference behavior.

## Activity A - Find the transferable pattern

Choose one hypothetical platform capability, such as:

- an application environment that composes a service, configuration, and
  observability resources;
- an optional backup or integration enabled only by a user choice; or
- a repeated resource created for each declared worker or component.

Map the capability to the game:

| Platform question | Game evidence |
|---|---|
| What is the user-facing API? | |
| What is composed from one instance? | |
| Which fields are user input versus derived status? | |
| Where are repetition and conditional inclusion represented? | |
| What shows that the desired state has converged? | |

## Activity B - Compare capabilities and operating models

Write a technical comparison. Include:

1. **Use case:** Which recurring composition rules can an RGD express
   through a single API?
2. **Abstraction boundary:** Which decisions should platform users control,
   and which defaults or constraints should the platform own?
3. **Complexity:** Which repeated structure can an RGD centralize?
   Which indirection and dependencies does that abstraction introduce?
4. **Operations:** Who owns RGD changes, compatibility, observability,
   reconciliation failures, cleanup, and support?
5. **Risk:** What isolation, authorization, quotas, secrets handling, and
   versioning would need review before production use?
6. **Alternative:** When would a Helm chart, an existing GitOps workflow, or
   a purpose-built operator be more appropriate?
7. **Evidence:** Which observed behavior supports the comparison, and
   which capabilities would require additional validation?

## Activity C - State the boundary accurately

Complete these statements in your own words:

- In this demo, Go is responsible for ...
- In this demo, kro is responsible for ...
- CEL in an RGD is used here to ...
- The CEL playground is useful for ..., but it does not ...
- This game is a poor example for teaching ...

The game stores desired state in a custom resource to demonstrate resource
composition and reconciliation. High-throughput application storage,
production authentication and transport security require separate designs.

## Deliverable

Submit a one-page technical comparison or a short verbal review.
Connect observed RGD behavior to an API capability, describe its
operating requirements, and compare it with alternative mechanisms.
