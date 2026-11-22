# J3 API

The machine-readable contract is `api/openapi.yaml` (78 paths). Generated TypeScript types live in
`web/src/lib/api-types.d.ts` and are regenerated with `bun run generate:types`. A CI test asserts
that every registered route in `internal/api/server.go` appears in the spec.

## Authentication

By default the server requires a bearer token for every route except health checks, the public
boards and registration. Tokens are server-side sessions stored in `sessions`.

```bash
# 1. Begin sign-in (password is verified against a bcrypt hash)
curl -X POST http://127.0.0.1:8787/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"tony@okami.j3","password":"demo","provider":"password"}'

# 2. Verify the six-digit MFA code (demoCode is only returned when J3_DEMO=1)
curl -X POST http://127.0.0.1:8787/api/v1/auth/verify \
  -H 'Content-Type: application/json' \
  -d '{"email":"tony@okami.j3","code":"123456"}'

# 3. Use the token on every subsequent request
curl -H 'Authorization: Bearer <token>' http://127.0.0.1:8787/api/v1/demo/admin/users
```

`POST /api/v1/auth/demo` issues a runner session for QA **only** when `J3_DEMO=1`. The local demo
scripts and Playwright configuration set this variable; production deployments must not.

## Boundary model

- **Public**: event details, public dashboard data, belt catalogue, kata catalogue, schedules,
  results, medal tables, technique palette and simulation.
- **Authenticated**: data explorer, account admin, permissions, check-in, imports, announcements,
  bracket generation and uploads.
- **Per-contest**: fight mutations additionally require the caller to be assigned to that contest
  and to hold a role with the `fights` capability.

Role–resource CRUD permissions are stored in `role_permissions` and editable by the runner at
`POST /api/v1/demo/permissions`. Every change is written to the audit log with the authenticated actor.

## Worked examples

### Create a tournament

```bash
curl -X POST http://127.0.0.1:8787/api/v1/tournaments \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"Autumn Open","disciplines":[{"name":"Continuous Fighting","points":[{"label":"Takedown","value":5}]}],"categories":[{"name":"Adults","kind":"age","min":18,"max":99,"ranges":[{"kind":"weight","min":60,"max":90}]}],"saveTemplate":true}'
```

### Generate pools, brackets and officials

```bash
curl -X POST http://127.0.0.1:8787/api/v1/demo/auto-run \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{}'
```

### Score a contest

```bash
curl -X POST http://127.0.0.1:8787/api/v1/fights/ok-fight-01/event \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"corner":"red","kind":"score","points":5,"technique":"Taoshi (Takedown)"}'
```

### Search techniques and configured points

```bash
curl 'http://127.0.0.1:8787/api/v1/demo/techniques?eventId=ev-okami&q=seoi&group=throws'
# Response includes the catalogue match, resolvedPoints and the event's pointOptions.
```

Pass a catalogue `techniqueId` (and optionally `pointLabel`) when scoring. The server resolves the
point value and display label from the event rules, so the values above are never trusted from the
request body.

### Call a technical result

```bash
curl -X POST http://127.0.0.1:8787/api/v1/fights/ok-fight-24/technical \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"winner":"blue","reason":"injury","detail":"shoulder injury"}'
```

The outcome is `win_blue` (or `dq_blue` for cheating/disqualification), the bout locks to
`forfeit`, and an audited technical score event is published to the live displays.

### Safeguarding

Registration of anyone under 16 requires `guardianName`, `guardianContact` and `guardianConsent`.
Guardian fields are stored on the fighter record and are not exposed by public endpoints.

## Errors

`400` validation, `401` unauthenticated, `403` forbidden, `404` not found, `409` conflict,
`500` generic with the details logged against the request ID. Raw database errors are never
returned to clients.
