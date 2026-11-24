# J3 (J Cubed)

A white-label Japanese Jiu Jitsu tournament creator and operator.

> **Read [AGENTS.md](AGENTS.md) first.** It is the living source of truth for scope,
> decisions, domain rules and roadmap. This README is only a quick-start.

## What this is

One configurable engine for any Jiu Jitsu event — any organiser, club, rule set,
discipline, category, format or display — fully branded and fully offline-capable.
Setup is deliberately guided and plain-language: **no technical knowledge required**.

## Stack

| Layer | Technology |
|---|---|
| Venue server | Go + SQLite (`modernc.org/sqlite`), REST + WebSocket |
| Web / mobile UI | SvelteKit + TypeScript, Capacitor for iOS/Android/tablets |
| Frontend tooling | **Bun** (`bun install`, `bun run`, `bun test`) — no npm |
| Shared types | OpenAPI 3.1 spec → generated TypeScript types |
| Tests | Go `testing`, Bun/Vitest + Playwright |

## Repository layout

```
AGENTS.md            Living project brief (read first)
api/openapi.yaml     Shared schema source of truth
cmd/j3d              Venue server binary
internal/            Go packages (config, db, api, auth, domain)
web/                 SvelteKit frontend + Capacitor
.github/workflows    CI
```

## Quick start — full local demo

```sh
bun run dev      # friendly TUI menu (emoji + spinner): run, QA, reset, build, deploy, mobile
bun run demo     # skip the menu and start the full seeded demo immediately
```

- UI: <http://localhost:5173>
- API: <http://127.0.0.1:8787/api/v1/demo>
- Reset the demo data: `bun run reset`
- Run all checks: `bun run check && bun run test && bun run build`
- Full local CI/CD: `bun run cicd`
- Release artefacts: `bun run release`
- Prove every TUI action: `bun run test:tui`

### Dedicated screens (each opens in its own tab)
- `/display` — full-screen live scoreboard with officials, big fighter cards, timer and celebrations
- `/public` — public dashboard with live fights, next-up callouts, leaderboard and results
- `/brackets` — elimination and round-robin progression with result icons and promotion
- `/tables` — overall/category tables, medals and kata panel scores
- `/pools` — pools and all matchups
- `/control` — open views on other screens, save dashboard templates, theme and PA announcements

The TUI menu also builds a **live executable** (`build/server/j3d`) and Capacitor targets for
Android and iOS. Local builds are staged in a single predictable tree:

```
build/android/   self-contained QA APK plus bundled seeded SQLite and API snapshot
build/ios/       iOS demo assets and, on macOS with Xcode, the simulator .app
build/web/       production web bundle
build/web-dev/   local development web bundle
build/web-qa/    QA/staging web bundle
build/web-prod/  production web bundle
build/server/    local server executables
```

`bun run build` stages the web bundle in `build/web`; `bun run build:dev`, `build:qa` and
`build:prod` (or `build:flavors`) stage explicit flavours with optional `J3_DEV_API_URL`,
`J3_QA_API_URL` and `J3_PROD_API_URL` API targets. `bash scripts/local-build.sh all` builds every
local target. `bun run release` still mirrors the web/server outputs into `dist/` for the existing
deploy flow.

Regenerate shared types after editing `api/openapi.yaml`:

```sh
bun run generate:types
```

## API explorer

A developer/admin API console is available at <http://localhost:5173/apidocs> after `bun run dev`.
It lists key endpoints, shows the OpenAPI contract and guide, and can send authenticated requests
from the browser. Open it from the TUI menu with **🔌 API explorer**.

## Mobile QA

```sh
bun run dev              # menu → Mobile → Build demo APK / Start Pixel emulator / Install & launch
bash scripts/mobile.sh build     # self-contained QA APK at build/android/j3-demo.apk
adb install -r build/android/j3-demo.apk
```

The default local QA APK needs no venue server: it bundles the seeded Okami SQLite file and a
read-only API snapshot, auto-logs in as admin, unlocks every screen on a phone for full end-to-end
QA and persists simulated writes locally. The drawer shows the build flavour. To exercise live LAN
sync instead, bake a server URL at build time:

```sh
J3_SERVER_URL="$(bash scripts/mobile.sh url)" bash scripts/mobile.sh build
```

GitHub Actions builds the Android APK and an iOS simulator .app on every push/PR; signed IPA
builds run when Apple secrets are configured.

## Status

Delivered white-label venue platform with a self-contained QA APK. See `AGENTS.md` §13 for decisions
and §14 for current release/hosting status.
