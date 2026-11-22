# CI/CD & release process

## Branches

`main` is protected. PRs run verification; release tags build and publish installable QA apps plus
production server/client artifacts.

## Automated gates (`.github/workflows/ci.yml`)

Go vet/test/race/build; Bun check/test/build; generated-type parity; smoke, load and Playwright + axe
E2E; workflow syntax/TUI checks. All must pass before merge.

## Mobile (`.github/workflows/mobile.yml`)

Every run:

1. regenerates the seeded Okami 2027 SQLite database and read-only API snapshot;
2. builds the web client with `VITE_OFFLINE_DEMO=1 VITE_QA_AUTOLOGIN=1` (first launch seeds the
   bundled data and signs in as the seeded admin);
3. produces an installable Android QA APK and an iOS simulator app;
4. boots an Android emulator and iOS simulator, launches each native client and uploads screenshots
   and logs as smoke evidence.

Artifacts: `j3-android-qa-<sha>`, `j3-android-smoke-<sha>`, `j3-ios-simulator-<sha>`,
`j3-ios-smoke-<sha>`. A real-device `j3-ios-qa-<sha>.ipa` is also produced when the Apple signing
secrets are configured; release builds fail loudly if they are missing.

Required iOS secrets: `APPLE_CERTIFICATE_P12`, `APPLE_CERTIFICATE_PASSWORD`,
`APPLE_PROVISIONING_PROFILE`, `APPLE_TEAM_ID`. Optional: `APPLE_PROVISIONING_PROFILE_NAME`,
`APPLE_EXPORT_METHOD` (defaults to `ad-hoc`).

## Releases (`.github/workflows/release.yml`)

Release builds call CI and mobile (requiring the signed iOS QA app), then package final production
binaries (`j3d` for Linux/macOS/Windows), the optimised web client bundle and `LICENSE` as
`j3-production-<sha>`. All artifacts are attached to the GitHub release.

## Environments (`.github/workflows/deploy.yml`)

Staging/production deploys are environment-gated and consume the release bundle. Hosted owner
credentials come from `J3_OWNER_*`; never commit passwords.

## Local rigour

Run `bun run cicd` or `bun run test:tui` before pushing. `bun run docs` rebuilds the documentation
site; `bun run sync:api-docs` keeps the contract/guides/search index and shipped licence in sync.
