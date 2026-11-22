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

## Data and reset

Use **Reset demo data** to restore the seeded tournament. The full seed includes 40 fighters, four clubs, kata scores, medals, pools, brackets and an audit trail. A single SQLite file is the portable backup.

## Troubleshooting

If the docs site shows a proxy error, confirm the API is listening on `127.0.0.1:8787` and that the Vite proxy regex forwards `^/api/`. If a page shows stale data after a reset, reload once — the client automatically retries with a fresh demo session after a 401.
