# J3 API

The machine-readable contract is `api/openapi.yaml` (83 paths, 87 operations). Generated TypeScript
types live in `web/src/lib/api-types.d.ts`; regenerate with `bun run generate:types`. Route parity is
tested in Go and CI.

## Authentication

Bearer sessions are stored server-side. Public routes are health checks, public boards and
registration. In demo mode (`J3_DEMO=1`):

```bash
curl -X POST localhost:8787/api/v1/auth/demo              # seeded admin session
curl localhost:8787/api/v1/demo/credentials              # accounts, password "demo"
curl -X POST localhost:8787/api/v1/auth/demo \
  -H 'Content-Type: application/json' -d '{"userId":"u-official"}'
```

Hosted bootstraps an owner from `J3_OWNER_*`; a generated one-time password is logged and
`POST /api/v1/auth/password` clears `mustChangePassword` after reset.

### Invitation chain

Self-signup creates only fighters/spectators. Admins invite all roles; runners invite officials,
club owners and participants; officials invite club owners and participants; club leaders invite
only their own club. Tokens are hashed and `/invite/<token>` URLs are returned immediately.

```bash
curl -X POST localhost:8787/api/v1/invites -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Nina","email":"nina@club.test","role":"official"}'
curl -X POST localhost:8787/api/v1/invites/accept -H 'Content-Type: application/json' \
  -d '{"token":"<token>","name":"Nina","password":"choose-a-password"}'
```

## Boundary model

- **Public**: event details, dashboard, belts/kata catalogues, schedule, results, medals, technique
  palette and simulation.
- **Authenticated**: query API, uploads, check-in, imports, announcements, brackets and templates.
  Sessions re-read status and role on every request; suspended accounts and stale roles lose access.
- **Per-contest**: fight mutations require assignment (admin/runner override); points are resolved
  from the rule snapshot, never trusted from the client.

## Worked examples

```bash
# Seeded data and generated draw
curl -H "Authorization: Bearer $TOKEN" localhost:8787/api/v1/demo
curl -X POST localhost:8787/api/v1/demo/pools/generate -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"eventId":"ev-okami","categoryId":"cat-adult-76"}'

# Score a technique (points resolved server-side)
curl -X POST localhost:8787/api/v1/fights/ok-fight-01/event -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"corner":"red","techniqueId":"tech-seoi-nage","pointLabel":"ippon"}'

# Technical result and audit
curl -X POST localhost:8787/api/v1/fights/ok-fight-24/technical -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"winner":"red","reason":"no_show"}'
```

## Errors

`400` validation, `401` missing/stale session, `403` outside authority, `404` unknown target,
`409` duplicate/conflict, `500` generic (details logged, not leaked).
