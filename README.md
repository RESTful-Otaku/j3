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

## Quick start

```sh
# Venue server
go run ./cmd/j3d --addr :8787 --db ./data/j3.db
curl http://localhost:8787/healthz

# Frontend (Bun only)
cd web && bun install && bun run dev
```

Regenerate shared types after editing `api/openapi.yaml`:

```sh
cd web && bun run generate:types
```

## Status

Phase 0 (foundation) in progress. See `AGENTS.md` §11 for the roadmap and §14 for decisions.
