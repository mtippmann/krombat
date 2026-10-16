# 01 - Observe the graph

**Goal:** Build a concrete model of what one platform-facing API instance
creates and how its resources converge.

## Activity A - Start from the API

1. Create a Dungeon with three monsters in the game UI. Note the Dungeon name.
2. Open the resource inspector. Record the Dungeon's `spec` and `status`
   separately. Find the custom resources, ConfigMaps, and Secrets related to
   this Dungeon.
3. If you have kubectl, inspect the same named instance and its labelled
   resources:

   ```bash
   kubectl --context minikube -n rpg-demo get dungeon team-workshop -o yaml
   kubectl --context minikube -n rpg-demo get hero,monster,boss,treasure,loot,cm,secret \
     -l game.k8s.example/dungeon=team-workshop
   ```

   On Docker Desktop use `--context docker-desktop` instead. Keep the instance
   name in the commands aligned with your own Dungeon.

### Questions

1. Which object is the user-facing input to the graph? Which fields describe
   requested game state, and which fields are reported status?
2. List the direct game-resource kinds composed by the Dungeon graph. Which
   objects are composed by those child resource graphs?
3. Which fields make the instance and its related resources easy to find or
   inspect? What value do printer columns add for a platform user?
4. Draw a simple dependency map: Dungeon -> game resources -> supporting
   resources. Mark any relationship you inferred from labels or graph
   references rather than Kubernetes owner references.

## Activity B - Observe convergence

1. Attack a monster once, then revisit the Dungeon and Monster resources.
2. Compare the saved Dungeon game fields with the projected Monster `spec` and
   the Monster graph's status.
3. Observe one complete reconciliation, then repeat the inspection after the
   UI has refreshed.

### Questions

1. What changed first in the requested game state? What did kro subsequently
   project or derive?
2. Are the related objects updated as one synchronous transaction? What does
   the observed sequence imply for a UI or client that reads during
   reconciliation?
3. Which parts of this path are an example of platform orchestration, and
   which parts are game-specific behavior?

## Activity C - Optional-resource lifecycle

1. Continue until one monster is defeated; inspect whether its Loot resource
   appears.
2. Defeat the boss and open the treasure; inspect the Treasure Secret only
   through metadata or the resource inspector. Do not copy or expose secret
   values.
3. Delete only your Dungeon through the UI and observe that its composed
   resources are cleaned up.

### Questions

1. Which conditions control whether Loot and the Treasure Secret exist?
2. What does the cleanup tell you about treating the Dungeon as an application
   or environment boundary?
3. Which aspects of kro's ownership and cleanup would you want documented in
   an internal platform runbook?

## Deliverable

Keep a one-page graph map with:

- the user-facing API and its important inputs;
- child and supporting resource kinds;
- one example of desired state, derived status, and an optional resource; and
- one operational question you would investigate before adopting this pattern.
