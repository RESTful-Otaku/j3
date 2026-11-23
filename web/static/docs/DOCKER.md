# Containerised QA demo

A full, repeatable J3 demo runs in containers: a seeded venue API, a static web front end with a
same-origin `/api` proxy (SSE-safe) and a deterministic tournament snapshot. Use Docker if available;
`scripts/qa-docker.sh` falls back to Podman automatically.

## Quick start

```bash
bash scripts/qa-docker.sh full          # or: bun run qa:docker:full
bash scripts/qa-docker.sh url
```

Default URLs are `http://localhost:5173`, `http://localhost:5173/simulate` and
`http://localhost:8787/healthz`. Override ports with `J3_API_PORT` and `J3_WEB_PORT`.

## Modes

| Command | What it boots |
|---|---|
| `up` / `full` | Full tournament plus mid-event Okami 2027 seed |
| `okami` | Okami 2027 mid-event seed only |
| `clean` | Minimal base seed with a 24-fighter snapshot |
| `light` | Same small seed/snapshot as `clean` |
| `dev` | Full seed with Vite hot reload mounted from `./web` |

Each mode recreates the `j3-qa-data` volume, so results are deterministic and disposable.

## Lifecycle

```bash
bash scripts/qa-docker.sh build      # build API + prod/dev web images
bash scripts/qa-docker.sh status     # container state
bash scripts/qa-docker.sh logs       # recent API logs
bash scripts/qa-docker.sh follow     # stream API logs
bash scripts/qa-docker.sh simulate   # regenerate /data/simulation.json
bash scripts/qa-docker.sh reset      # discard data volume
bash scripts/qa-docker.sh down       # remove containers
```

`J3_SIM_SEED` and `J3_SIM_FIGHTERS` control the generated replay. The API image also exposes
`j3d simulate -seed 42 -fighters 50 -out snapshot.json` for snapshot generation without curl/jq.

## Compose equivalent

When a Compose provider is installed, `compose.yaml` is equivalent to the default `full` mode:

```bash
docker compose up --build
# or: podman-compose up --build
```

The script is preferred locally because it works with Podman directly and needs no Compose provider.
The multi-stage web image pins nginx so `/api` is proxied to the venue API on the same origin; live
SSE and the global replay widget therefore work from every display URL.

## Replays and medals

Open `/simulate` to switch the whole app into **Simulated tournament** mode. The persistent replay
widget follows every view and supports play/pause, restart, bout stepping, scrubbing and speed. The
nav toggle returns to the real venue state at any time. Simulated fights use the catalogue point
values (typical full totals 10–40) and instant-win submissions.

Seeded and recomputed medals are keyed by event, category, discipline and event type. Fighting,
Ne-Waza, Randori, continuous/sparring and kata each receive their own labelled podium; dual-bronze
policies are honoured. `POST /api/v1/demo/medals/recompute` rebuilds the full deterministic set
from completed results without overwriting manual overrides.
