# CI/CD & release process

## Branches

`main` is protected. PRs run verification; release tags build artifacts.

## Automated gates (`.github/workflows/ci.yml`)

Go vet/test/race/build; Bun check/test/build; generated-type parity; OpenAPI route parity; smoke
suite; Playwright + axe E2E; workflow syntax checks. All must pass before merge.

## Mobile (`.github/workflows/mobile.yml`)

Builds an Android APK and iOS simulator app. Mobile builds regenerate icons, seed a mid-event Okami
2027 database, capture a read-only API snapshot and set `VITE_OFFLINE_DEMO=1`. Artifacts are uploaded
for QA.

## Environments (`.github/workflows/deploy.yml`)

Staging/production deploys are environment-gated. Hosted owner credentials come from `J3_OWNER_*`;
never commit passwords.

## Releases (`.github/workflows/release.yml`)

Tag-driven release builds attach server binaries and mobile artifacts. Signed IPA needs Apple secrets.

## Local rigour

Run `bun run cicd` or `bun run test:tui` before pushing. `bun run docs` rebuilds the documentation
site; `bun run sync:api-docs` keeps the contract/guides/search index in sync.
