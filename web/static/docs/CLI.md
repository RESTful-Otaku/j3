# CLI & TUI reference

`./scripts/dev.sh` is the main operator menu. It uses plain-language choices, emoji and arrow-key navigation. Press `q` or `Esc` to step back to the parent menu.

## Start here

| Menu action | What it does |
|---|---|
| Run local demo | Boots the venue server and web UI with the seeded Okami tournament |
| Run live executable | Builds and starts the single Go binary with full seed data |
| Open docs & guides | Starts the local services if needed and opens `/docs` |
| Open API explorer | Opens the live OpenAPI explorer at `/docs/api` |

You can also run `bun run docs` from the repository root to open the documentation site directly.

## QA and verification

| Menu action | Command | Purpose |
|---|---|---|
| Smoke tests | `bun run smoke` / `smoke:full` | API and UI assertion sweeps |
| Browser tests | `bun run e2e` | Playwright journeys and axe accessibility scans |
| Load test | `bun run load` | 200 concurrent requests |
| Go tests | `go test -race ./...` | Backend unit, property and race checks |
| Check workflows | `bun run check:workflows` | CI YAML validation |

## Build and deploy

The **Build** menu creates a release binary and can locally verify the CI/CD script. Deployment is environment-gated: staging and production require the matching environment variables and credentials. Release artifacts include the venue binary and mobile builds.

## Mobile

The **Mobile** menu prepares Capacitor projects, builds an Android APK and launches an Android Studio emulator. On this host the reliable emulator command uses `~/Android/Sdk`, the `j3-pixel-10-pro` AVD and `-gpu auto`. iOS builds require macOS and Xcode.

For a locally testable seeded build:

```bash
bash scripts/mobile.sh build      # standalone, offline-capable APK
bash scripts/mobile.sh install    # install on the connected emulator/device
```

Every mobile build regenerates the J3 launcher/splash icons from `web/static/brand/logo.png` via `scripts/mobile-icons.py`, then runs `scripts/mobile-demo-data.sh`. That script seeds a fresh mid-event Okami 2027 SQLite file (`web/static/j3-demo.db`) and a read-only API snapshot (`web/static/j3-demo-data.json`) from a temporary venue server. Both are bundled inside the APK. When the phone can reach the venue LAN, the app uses live data; when it cannot, public/participant/club/guardian screens fall back to the bundled seeded snapshot, so the tournament can be demonstrated anywhere.

To bake a specific venue URL for live scoring QA, start the venue server from the TUI (**Run demo for phone QA**) and build with:

```bash
J3_SERVER_URL="$(bash scripts/mobile.sh url)" bash scripts/mobile.sh build
```

The local demo server allow-lists Capacitor's `http://localhost` and `capacitor://localhost` origins. Phones intentionally show only public, participant, club and guardian screens; scoring, setup, audit and control-room tools remain available on tablet/desktop/display screens. Verified emulator evidence is saved under `docs/qa/`: `pixel-10-pro-icon.png` (branded launcher icon), `pixel-10-pro-running.png`/`pixel-10-pro-logcat.txt` (live LAN build) and `pixel-10-pro-offline.png`/`pixel-10-pro-offline-logcat.txt` (venue server stopped, bundled snapshot in use).

## Data and reset

Use **Reset demo data** to restore the seeded tournament. The default Okami Jiu Jitsu Challengers Tournament 2027 seed is deliberately **mid-event**: completed and drawn bouts, running and paused clocks, technical results, upcoming contests, partly judged kata, live schedule drift, medals and an audit trail. The full seed additionally includes a six-round bracket and deterministic 50-participant simulation data. It has 40 fighters, four clubs, kata scores, medals, pools and brackets. A single SQLite file is the portable backup.

## Troubleshooting

If the docs site shows a proxy error, confirm the API is listening on `127.0.0.1:8787` and that the Vite proxy regex forwards `^/api/`. If a page shows stale data after a reset, reload once — the client automatically retries with a fresh demo session after a 401.
