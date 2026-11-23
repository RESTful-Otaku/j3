# CI/CD & release process

## Branches

`main` is protected. PRs run verification; release tags build and publish versioned, installable
mobile clients plus production server/client artifacts.

## Free-only hosted CI

No billing account, payment method or spending limit is required. Every job in every workflow has a
job-level guard:

```yaml
if: github.event.repository.visibility == 'public'
```

GitHub-hosted runners are free for public repositories, and this repository is public, so the
hosted gates run normally at no cost. The visibility guard remains as a safety rail: if the repo is
ever made private again, every job is **skipped before a runner is requested** (including the 10×
`macos-15-intel` mobile job), so no billing method or spending limit is ever needed.
`scripts/check-workflows.py` fails locally if any job loses that gate.

Do not enable billing or a spending limit as a workaround for runner limits. Keep the repository
public for hosted CI, or move builds to self-hosted runners later if privacy is required. Local
gates and the containerised QA demo remain available regardless of visibility.

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

Release builds call CI and mobile, then attach exactly four download categories to the GitHub
release: GitHub's auto-generated **Source code** archives, one hosted server executable
(`j3-server-linux-amd64`), the production Android client (`j3-android-<version>.apk`) and the iOS
simulator client (`j3-ios-<version>.app.zip`, containing the `.app` bundle). No QA apps, docs bundles,
web archives or IPA files are published. The deploy workflow consumes the same server executable
artifact (`j3-server-release-<sha>`).

## Environments (`.github/workflows/deploy.yml`)

Staging/production deploys are environment-gated and consume the release bundle. Hosted owner
credentials come from `J3_OWNER_*`; never commit passwords.

## Local rigour

Run `bun run cicd` or `bun run test:tui` before pushing. `bun run qa:docker:full` boots the full
seeded stack through a working Docker/Podman daemon (see `docs/DOCKER.md`). `bun run docs` rebuilds
the documentation site; `bun run sync:api-docs` keeps the contract/guides/search index and shipped
licence in sync.
