# CLI & TUI reference

## Start here

`dev.sh` is the plain-language menu (arrows/Enter, q/Esc to go back). Numeric shortcuts work too.
`bun run dev` boots the seeded full-stack demo; `bun run api` starts just the venue server.

## QA and verification

`bun run smoke` (80 API/UI checks), `bun run smoke:full` (full-seed), `bun run e2e` (browser/axe),
`bun run load` (200 concurrent requests), `bun run test:tui` (57 menu actions).

## Build flavours

| Flavour | Web | Android | iOS assets | Backend |
|---|---|---|---|---|
| dev (local) | `bun run build:dev` | `bun run build:android:dev` | `bun run build:ios:dev` | local venue server (`J3_DEV_API_URL` or `scripts/mobile.sh url`) |
| QA (seeded, standalone) | `bun run build:qa` | `bun run build:android:qa` | `bun run build:ios:qa` | bundled demo database/API snapshot, no server needed |
| production/live | `bun run build:prod` | `J3_SERVER_URL=… bun run build:android:prod` | `J3_SERVER_URL=… bun run build:ios:prod` | public API origin baked from `J3_SERVER_URL`/`J3_PROD_API_URL` |

`bun run build:flavors`, `build:android:flavors` and `build:ios:flavors` produce all three at once.
For the mobile dev flavour, start the TUI's venue-server-for-phone-QA action (binds `0.0.0.0:8787`)
or set `J3_DEV_API_URL` explicitly; `scripts/mobile.sh url` prints the default LAN address.
Only QA autologs in as admin and unlocks full phone routes; dev and production keep participant-only
phone scope. Production mobile builds require a real `J3_SERVER_URL` and are debug-signed locally;
hosted release builds publish the signed production APK/`.app`.

## Build and deploy

`bun run build` builds Go + web + docs; `bun run release` produces release artifacts;
`bun run cicd` runs the local CI/CD sequence; `bun run deploy` is environment-gated. Production web
is mirrored into `build/web` and `dist/web`.

## Mobile

`bash scripts/mobile.sh build` is the QA alias (standalone seeded APK). `install` starts the Pixel 10
Pro AVD (Android Studio SDK, `-gpu auto`) and installs/launches it. `url` prints the LAN URL used by
the dev flavour. Staged APKs live in `build/android/` as `j3-dev.apk`, `j3-qa.apk` (also
`j3-demo.apk`) and `j3-prod.apk`.

## Containerised QA

The **🐳 Docker QA ▸** TUI menu runs the container stack in `full`, `okami`, `clean`, `light` or `dev`
mode. The same commands are available as `bun run qa:docker*` (see `docs/DOCKER.md`).
`j3d simulate -seed 42 -fighters 50 -out snapshot.json` writes a deterministic tournament snapshot
without needing curl or a running server.

## Data and reset

`bun run reset` resets demo data; `bun run seed:full` creates the complete deterministic tournament.
`GET /api/v1/demo/credentials` lists seeded accounts (password `demo`).

## Troubleshooting

If the emulator crashes, check that the Android Studio SDK and AVD `j3-pixel-10-pro` are selected.
If the venue URL changes, rebuild with `J3_SERVER_URL=...`. Logs and screenshots live in `docs/qa/`.
