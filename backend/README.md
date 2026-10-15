# Local demo backend

The backend runs inside minikube in `rpg-system` and accesses only game
resources in `rpg-demo`. It uses upstream kro **v0.9.4**, without a fork or
write-back extensions. Go owns deterministic turns in typed `spec.game`;
kro reconciles child resources and projects derived status. Clients poll.

`SESSION_SECRET` must be generated during local setup and contain at least
32 non-padding bytes. Startup reports invalid configuration explicitly.
`PORT` defaults to `8080`. There are no OAuth, test-login, WebSocket,
leaderboard, or telemetry routes, and no kubeconfig fallback.

## HTTP API

All responses are JSON except successful logout/delete (`204`). Errors use
`{"error":"message"}`. Dungeon operations require the signed demo session.

| Method | Path | Request / response |
| --- | --- | --- |
| GET | `/healthz` | `{"status":"ok"}`, no authentication |
| POST | `/api/v1/auth/login` | No body; `{"login":"demo","avatarUrl":""}` and session cookie |
| GET | `/api/v1/auth/me` | `{"login":"demo","avatarUrl":""}` or `401` |
| POST | `/api/v1/auth/logout` | Clears session cookie; `204` |
| GET | `/api/v1/dungeons` | Raw `Dungeon[]`, filtered to demo-owned `rpg-demo` objects |
| POST | `/api/v1/dungeons` | `{name,monsters,difficulty,heroClass?,namespace?}`; raw Dungeon, `201` |
| GET | `/api/v1/dungeons/{namespace}/{name}` | Raw Dungeon |
| DELETE | `/api/v1/dungeons/{namespace}/{name}` | `204`, with UID/resource-version preconditions |
| POST | `/api/v1/dungeons/{namespace}/{name}/attacks` | `{target,seq,damage?}`; raw updated Dungeon; `target:"open-treasure"` opens the chest |
| POST | `/api/v1/dungeons/{namespace}/{name}/actions` | Compatibility alias: `{action:"open-treasure",seq}`; raw updated Dungeon |
| GET | `/api/v1/dungeons/{namespace}/{name}/resources` | Raw Kubernetes object array, parent first |
| POST | `/api/v1/dungeons/{namespace}/{name}/cel-eval` | `{expr}`; `{result:string,error?:string}` |

The cookie is signed, HttpOnly, SameSite=Lax, and deliberately not Secure
because this demo uses local HTTP. Sessions expire after four hours.
Login and CEL evaluation are rate-limited; `429` includes `Retry-After`.

Creation accepts DNS-label names (1–63 characters), 1–10 monsters, difficulty
`easy|normal|hard`, and only `warrior` (the default). Namespace defaults to
`rpg-demo`; other namespaces are rejected. Initial HP is hero **200**,
monsters **30/50/80**, and boss **200/400/800**.

Combat targets must exactly match `{name}-monster-{zero-based-index}` or `{name}-boss`.
Hero damage is **25/35/45**; client damage is ignored. All surviving monsters
counter after a monster attack with **1/2/3** damage each; apply the warrior's
25% reduction to their sum and round down. Boss counters use **3/5/8** damage,
scaled by phase **10/13/16 ÷ 10**, then reduced 25%, rounding down at each step.
Phases use the post-hit boss HP: above 50%, above 25%, then 25% or below.
HP never drops below zero. The boss stays locked until every monster is dead;
dead targets and terminal combat return explicit `400` errors. Treasure opens
once, only after boss defeat while the hero is alive.

Both mutations require the current `spec.attackSeq` as `seq` and increment
that same sequence. A stale sequence or Kubernetes update conflict returns
`409`, without automatic retries. Only `spec.game` is authoritative; stale
`status` is never used to calculate turns.

The resource inspector only lists allowlisted game CRs, ConfigMaps, and
Secrets bearing this dungeon's `game.k8s.example/dungeon` label. It follows
actual upstream `kro.run/instance-id` ancestry, validating each immediate
parent's name, namespace, kind, group, and version. Matching names or a
dungeon label alone cannot expose arbitrary resources.

The CEL playground uses upstream `BaseDeclarations()` with actual Dungeon
`schema.spec`, `schema.status`, and `schema.metadata` bindings. Evaluation is
read-only, limited to 2,000-byte expressions, a 64-level parser recursion
limit, 10,000 evaluation cost units, and a 100 ms evaluation deadline.
Compile/evaluation errors return `200` with `result:""` and `error`.

## Container validation

Run from the repository root:

```sh
docker build --target test -t krombat-backend-test backend
docker build -t krombat-backend:local backend
```

The test target runs unit tests and `go vet` using Go **1.26.3**. The final
image also depends on this test stage and runs as a non-root user. Tests
use a fake dynamic Kubernetes client; no cluster access is required.
