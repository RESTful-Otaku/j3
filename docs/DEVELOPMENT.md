# Developer guide

## Stack

Go single binary with pure-Go SQLite (`modernc.org/sqlite`), REST + SSE, OpenAPI contract;
SvelteKit + TypeScript + Bun (never npm); Capacitor for iOS/Android. Types are generated into
`web/src/lib/api-types.d.ts`.

## First run

```bash
bun run dev          # venue server + web UI with seeded Okami 2027 data
bun run demo:full    # full deterministic tournament seed
bash scripts/dev.sh  # TUI menus for QA/build/mobile/docs
```

## Project layout

`cmd/j3d` server; `internal/{api,auth,config,db,rules,seed,simulator,techniques,units}` backend;
`api/openapi.yaml` contract; `web/` SvelteKit app; `scripts/` QA/build helpers; `docs/` guides.

## API and generated types

Change `api/openapi.yaml`, then `bun run generate:types`. A Go test asserts route parity; a CI step
fails if the checked-in types are stale. Docs indexes are regenerated with `bun run sync:api-docs`.

## Adding a screen

Use existing route/lib patterns, keep role/phone guards in `web/src/lib/nav.ts`, add an E2E check,
and keep W/L/D accessible with glyphs. Operator tools belong on tablet/desktop, not phones.

## Live scoring rules

Clients send `techniqueId`/`pointLabel`; the server resolves points from the rule snapshot, enforces
assignment and writes committed events. Visualisations query committed data and receive SSE
updates.

## Verification

`go vet ./...`, `go test ./...`, `go test -race ./...`, `bun run check`, `bun run test`,
`bun run e2e`, `bun run build`, `bun run smoke`, `bun run load`, `bun run test:tui`,
`bun run check:workflows`.

## Troubleshooting

If API calls 401 in demo, the session may be stale — the client retries once with a fresh session.
For mobile CORS use local demo origins `http://localhost` / `capacitor://localhost`. Check
`docs/API.md` for auth and boundary examples.
