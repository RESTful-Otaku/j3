# CI/CD & release process

## Branches

`main` is protected. PRs run verification; release tags build and publish versioned, installable
mobile clients plus production server/client artifacts.

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

QA artifacts: `j3-android-qa-<sha>`, `j3-android-smoke-<sha>`, `j3-ios-simulator-<sha>`,
`j3-ios-smoke-<sha>`.

Tagged releases also build **hosted production clients** with the public `J3_SERVER_URL` baked in:

- `j3-android-<version>.apk` — signed release APK (stable key when Android secrets are set);
- `j3-ios-<version>.app.zip` — unsigned simulator app for Xcode/simulator testing;
- `j3-ios-<version>.ipa` — real-device ad-hoc/development IPA when Apple secrets are configured.

Both production clients get their own emulator/simulator smoke run. Signing credentials are optional
so a release without Apple/Android secrets still completes; see `docs/ENVIRONMENT.md`.

## Releases (`.github/workflows/release.yml`)

Release builds call CI and mobile, then package final production binaries (`j3d` for Linux
amd64/arm64, macOS arm64 and Windows), the optimised web client bundle, the OpenAPI reference, full
guide/documentation site and `LICENSE`. Downloads are `j3-production-<sha>.tar.gz`; the deploy
workflow consumes the companion `deploy-production-<sha>` artifact. Only curated `j3-*` release
assets are attached to the GitHub release, avoiding GitHub's secondary rate limit from uploading
every static file separately.

## Environments (`.github/workflows/deploy.yml`)

Staging/production deploys are environment-gated and consume the release bundle. Hosted owner
credentials come from `J3_OWNER_*`; never commit passwords.

## Local rigour

Run `bun run cicd` or `bun run test:tui` before pushing. `bun run docs` rebuilds the documentation
site; `bun run sync:api-docs` keeps the contract/guides/search index and shipped licence in sync.
