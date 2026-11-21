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

### Dedicated screens (each opens in its own tab)
- `/display` — full-screen live scoreboard with officials, big fighter cards, timer and celebrations
- `/public` — public dashboard with live fights, next-up callouts, leaderboard and results
- `/brackets` — elimination and round-robin progression with result icons and promotion
- `/tables` — overall/category tables, medals and kata panel scores
- `/pools` — pools and all matchups
- `/control` — open views on other screens, save dashboard templates, theme and PA announcements

The TUI menu also builds a **live executable** (`bin/j3d` + `dist/web`) and Capacitor targets
for Android and iOS.

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
bash scripts/mobile.sh url       # LAN URL for a physical phone on the same Wi-Fi
bash scripts/mobile.sh build     # builds artifacts/j3-demo.apk
```

For a physical phone, use the menu option **Run venue server for phone** (binds all interfaces
with demo auth), then install `artifacts/j3-demo.apk` via `adb install -r` or by opening it on the
device. The venue URL is baked into the APK at build time with `J3_SERVER_URL`.

GitHub Actions builds the Android APK and an iOS simulator .app on every push/PR; signed IPA
builds run when Apple secrets are configured.

## Status

Phase 0 (foundation) in progress. See `AGENTS.md` §11 for the roadmap and §14 for decisions.
