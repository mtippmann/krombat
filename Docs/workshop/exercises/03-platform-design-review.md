# 03 - Platform design review

**Goal:** Decide whether the patterns in this game belong in a small company's
internal platform, and identify the conditions required to use them safely.
This is a design exercise; it does not ask you to modify code or deploy
resources.

Assume a small company has a platform team of a few engineers and product teams
that need repeatable environments. The company already runs Kubernetes and
wants a self-service API that composes an application with supporting
resources. Use the current game as a reference, not as a production template.

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

## Activity B - Make an adoption recommendation

Write a short recommendation to the engineering manager. Include:

1. **Use case:** What recurring platform problem would an RGD solve better
   than asking every team to assemble resources themselves?
2. **Abstraction boundary:** Which decisions should platform users control,
   and which defaults or constraints should the platform own?
3. **Complexity test:** What would make an RGD simpler than the team's current
   templates or deployment workflow? What would make it unnecessary
   indirection?
4. **Operations:** Who owns RGD changes, compatibility, observability,
   reconciliation failures, cleanup, and support?
5. **Risk:** What isolation, authorization, quotas, secrets handling, and
   versioning would need review before production use?
6. **Alternative:** When would a Helm chart, an existing GitOps workflow, or
   a purpose-built operator be more appropriate?
7. **Decision:** Recommend a small pilot, defer adoption, or use a different
   approach. State one measurable signal that would change your decision.

## Activity C - State the boundary accurately

Complete these statements in your own words:

- In this demo, Go is responsible for ...
- In this demo, kro is responsible for ...
- CEL in an RGD is used here to ...
- The CEL playground is useful for ..., but it does not ...
- This game is a poor example for teaching ...

Avoid proposing that a production company place high-throughput application
state in Kubernetes just because this game stores desired game state in a
custom resource. The demo's local shared identity and HTTP endpoint are not
security patterns to copy.

## Deliverable

Submit a one-page recommendation or a short verbal review. A strong answer
connects an observed RGD pattern to a real platform need, names the costs and
owners, and makes a bounded recommendation rather than advocating kro by
default.
