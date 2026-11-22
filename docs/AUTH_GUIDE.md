# Accounts, roles and invitations

One identity system serves the web app, venue server and native phone build. Authority flows down;
privileged actions are audited.

## Demo and QA sign-in

Local/native demo builds seed accounts sharing password `demo`. Use **Demo accounts — one click to
test any role** on `/login`, or the searchable list at `/demo-accounts`. Switching mints a real
server session for that user, so role guards and private/public boundaries are genuine.
`GET /api/v1/demo/credentials` supplies the sheet; the APK bundles it for offline identity
switching. Writes still require the venue LAN server.

## Hosted owner

First hosted boot creates the admin (`j3@local` / `J3 Owner` by default). Override with
`J3_OWNER_EMAIL` / `J3_OWNER_NAME`. `J3_OWNER_PASSWORD` wins; otherwise a strong one-time password
is printed once in the startup log and `must_change_password` forces a reset at `/login` via
`POST /api/v1/auth/password`.

## Invite chain

Public signup creates only fighters and spectators. Everyone else is invited:

| Inviter | May invite |
|---|---|
| Admin | every role, including other admins |
| Runner | officials, club owners, fighters, spectators |
| Official | club owners, fighters, spectators |
| Club leader | fighters/spectators **in their own club only** |

Create an invite from **Admin → Invite an account** or **Club → Invite member**. The response
contains a token and `/invite/<token>` URL; the invitee sets their own password on acceptance.
Tokens are hashed at rest, can be revoked, and create/accept/revoke are audited. Password reset
never activates a pending privileged account.

## Enforced boundaries

- Roles come from the database on every privileged request; client-supplied roles are ignored.
- `admin` is stored as a flagged runner but exposed as `admin` in UI, API and audit.
- Club owners cannot change tournament templates, permissions or other clubs.
- Scoring requires contest assignment (admin/runner override); points remain server-resolved.
