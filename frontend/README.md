# Local vanilla kro demo frontend

The pixel-art UI plays a single-room Warrior dungeon and inspects actual composed
Kubernetes objects. Go computes turns and writes `Dungeon.spec.game`; upstream kro
composes the six resource graphs and aggregates their status. There is no
WebSocket, GitHub OAuth, inventory, progression, or custom kro write-back feature.

## Build

From the repository root:

```sh
docker build -t krombat-frontend:local frontend
```

The image runs `npm ci`, then `tsc -b && vite build`, and serves static assets on
port 3000 as an unprivileged nginx user. No external fonts, telemetry, or assets
are requested. The existing pixel-art assets are retained.

## Same-origin routing

Open the demo through its local ingress, not the frontend container directly.
The ingress must route `/api` to the backend in `rpg-system` and `/` to the
frontend. nginx does not proxy API calls, require HTTPS, add HSTS, or connect to a
production hostname. The UI uses only relative `/api/v1` URLs and the HttpOnly
session cookie set by the backend.

Expected routes:

- `GET /api/v1/auth/me`, `POST /api/v1/auth/login`, `POST /api/v1/auth/logout`
- `GET /api/v1/dungeons`: raw `Dungeon[]`
- `POST /api/v1/dungeons`: `{name, monsters, difficulty, heroClass: "warrior"}`
- `GET` / `DELETE /api/v1/dungeons/rpg-demo/<name>`
- `POST .../attacks`: `{target: "<name>-monster-<zero-based-index>" | "<name>-boss" | "open-treasure", seq}`
- `GET .../resources`: raw Kubernetes object array, restricted to the selected
  Dungeon and its owned demo resources, including demo loot/treasure Secrets
- `POST .../cel-eval`: `{expr}`, returning `{result: string, error?: string}`

The collection is polled every 3 seconds, the selected Dungeon every 1.5 seconds,
and its resources every 2 seconds. Mutations return the updated raw Dungeon.
Sequence conflicts are displayed as HTTP 409 and refresh state without replaying
the action. Boss locking, defeat, and victory use `spec.game`, never stale status.

Projected phase, monster count, and boss state are displayed separately with a
reconciliation warning when they differ from the saved game. Phase expectations
use the demo's 50% / 25% thresholds. The resource inspector displays unmodified
JSON snapshots, including controller conditions and resource versions.
Upstream kro tracks membership through ApplySet inventory and kro finalizers,
not default ownerReferences. The `kro.run/instance-id` label names the immediate
parent UID; the inspector does not invent owner-reference relationships.

CEL examples use the backend's `schema` activation, such as
`schema.spec.game.monsterHP.filter(hp, hp > 0).size()`. Evaluation is read-only and
results are explicitly snapshots rather than live or in-controller traces.
