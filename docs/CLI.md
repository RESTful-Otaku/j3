# CLI & TUI reference

## Start here

`dev.sh` is the plain-language menu (arrows/Enter, q/Esc to go back). Numeric shortcuts work too.
`bun run dev` boots the seeded full-stack demo; `bun run api` starts just the venue server.

## QA and verification

`bun run smoke` (80 API/UI checks), `bun run smoke:full` (full-seed), `bun run e2e` (browser/axe),
`bun run load` (200 concurrent requests), `bun run test:tui` (57 menu actions).

## Build and deploy

`bun run build` builds Go + web + docs; `bun run release` produces release artifacts;
`bun run cicd` runs the local CI/CD sequence; `bun run deploy` is environment-gated.

## Mobile

`bash scripts/mobile.sh build` regenerates J3 icons, seeds the bundled database/snapshot and builds
the APK. `install` starts the Pixel 10 Pro AVD (Android Studio SDK, `-gpu auto`) and installs/launches
it. `url` prints the baked LAN URL.

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
