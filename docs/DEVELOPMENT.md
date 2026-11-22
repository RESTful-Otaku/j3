# Developer guide

J3 is one Go binary (the venue server) plus one SvelteKit app (every screen). The binary owns
SQLite and all authoritative state; the web app is a set of URL-addressable views that read and
write through the API. No separate state is kept in displays.

## Stack

| Layer | Choice | Why |
|---|---|---|
| Server | Go 1.27, `modernc.org/sqlite` | Single portable binary, no CGO, one-file backup |
| API | REST + SSE, OpenAPI 3.1 in `api/openapi.yaml` | Shared contract and generated TypeScript types |
| Web | SvelteKit + TypeScript + Bun | One app, many display modes |
| Mobile | Capacitor (Android/iOS) + PWA | Judge phones and tablets use the same UI |
| Tests | Go tests, Vitest, Playwright/axe, smoke and load scripts | Correctness from rules engine to browser |

## First run

```bash
bun run dev          # interactive TUI: demo, QA, builds, mobile, docs
bun run demo         # non-interactive: seeded venue server + UI at :5173
bun run demo:full    # same, with a complete tournament (pools, fights, kata, medals)
```

Prefer `bun run dev` for exploration. The TUI has nested menus: arrow keys move, Enter selects,
`q`/Esc goes back, and number keys jump directly to an item. Choose **Docs & guides** there to
open this site; it also regenerates the API index first.

The demo binds the API to `127.0.0.1:8787` and Vite to `0.0.0.0:5173`. Vite proxies `/api/`,
`/healthz` and `/readyz` to the API. The regex-style proxy keys are deliberate: a plain `/api`
prefix also matches `/apidocs` and would proxy the documentation site to the Go server.

For phone QA use **Mobile → Run venue server for phone** (`bun run demo:phone`). It binds all
interfaces and sets a narrow CORS allow-list. The APK bakes the LAN URL into the JS bundle via
`VITE_API_BASE_URL` while the WebView loads local assets.

## Project layout

```text
api/openapi.yaml          Single source of truth for the HTTP contract
cmd/j3d                   Venue server entry point
internal/api              Routes, middleware, RBAC guards, SSE hub
internal/auth             Sessions, bcrypt/MFA, rate limits
internal/config           Tenant config snapshots and themes
internal/db               SQLite migrations and queries
internal/rules            Eligibility, seeding, brackets, tie-breaks, third place
internal/seed             Okami demo and deterministic full-tournament seed
internal/simulator        Reproducible end-to-end simulation
internal/units            Imperial → metric conversion
scripts/                  TUI, smoke, load, TUI harness, docs generation
web/src/routes            App screens, including /docs
web/src/lib               API client, shared UI, markdown renderer
docs/                     Human-readable guides and this developer documentation
```

## API and generated types

Change the contract in `api/openapi.yaml`, then run:

```bash
bun run generate:types
```

This writes `web/src/lib/api-types.d.ts`, which is committed. CI regenerates it and fails if the
checked-in file is stale. A Go test also checks that every registered route appears in the spec.

The docs site reads a generated index at `web/static/docs/operations.json`. Regenerate the
reference, search index and markdown copies with:

```bash
bun run sync:api-docs
```

`scripts/build-docs-index.py` derives endpoint groups, parameters and example request bodies from
the OpenAPI spec. Add a new markdown file to `docs/` and an entry to `docs/manifest.json` to add a
page to the docs sidebar — the sync script copies it and includes it in search.

## Adding a screen

1. Create `web/src/routes/<name>/+page.svelte`. Use existing `app.css` tokens rather than raw
   colours so white-label themes still work.
2. If the screen is role-restricted, add it to `NAV_ITEMS` in `web/src/lib/nav.ts`. Public and docs
   routes are allow-listed in the root layout guard.
3. Read data through `web/src/lib/api.ts`, never from a separate store. Displays must reflect the
   committed database state.
4. Add a Playwright journey in `web/e2e/` and an axe scan with the shared helper.
5. Verify at phone, tablet portrait, tablet landscape and ultrawide sizes.

## Live scoring rules

The server owns the clock and assigns a global order to committed events. Judge devices send
idempotent mutations; a scorer must be assigned to the contest and hold the right role. Results
lock after confirmation, undo is recorded, and SSE fans state out to every display. Do not add a
client-side score buffer — it would break the audit trail.

## Verification

```bash
go vet ./...
go test ./... -count=1
go test -race ./...
bun run check                  # svelte-check
bun run test                   # Vitest
bun run smoke:full             # 70 API/UI assertions
bun run load                   # 200 concurrent requests
bun run e2e                    # Playwright + axe (run bun run e2e:install once)
bun run build                  # Go build + docs sync + static web build
bun run check:workflows
bun run test:tui               # drives all TUI menu actions
```

CI runs all of these on push and pull request. See [CI & releases](/docs/ci) for the workflow
details and release steps.

## Troubleshooting

| Symptom | Check |
|---|---|
| Docs site shows API JSON | Vite proxy key must be `^/api/`, not `/api`; restart Vite after editing `vite.config.ts` |
| `401` from a protected page | Refresh the demo session; demo identity switching mints a real bearer token |
| `500` from an upload or invite | Check `kind` on categories (defaults to `combined`) and upload capability scopes |
| Empty tables after a DB reset | Re-run `bun run reset && bun run seed` or the full demo; sessions self-heal on one 401 |
| Android WebView cannot reach API | Use `bun run demo:phone`; rebuild the APK with `VITE_API_BASE_URL` set to the LAN URL |
| Stale generated types | `bun run generate:types` and commit `web/src/lib/api-types.d.ts` |
