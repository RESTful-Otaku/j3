# CI/CD & release process

## Branches

`main` is protected: all changes arrive through a pull request. Use a prefixed branch:

| Prefix | Purpose |
|---|---|
| `feature/` | new capability |
| `bug/` | defect |
| `fix/` | correction |
| `chore/` | maintenance |
| `perf/` | performance work |
| `improv/` | UI/UX refinement |
| `release/` | release preparation |
| `hotfix/` | urgent production fix |

`develop` is optional integration branch. `ci.yml` validates the branch prefix on PRs.

## Automated gates (`.github/workflows/ci.yml`)

Every push/PR runs:
- **Go**: `go vet`, tests (unit/property/simulation), race detector, binary build.
- **Web (Bun)**: type check, Vitest, production build.
- **Smoke**: boots the real venue server and verifies 49 API/workflow/UI assertions.
- **required-checks**: aggregate gate; PR cannot merge unless all pass.

## Mobile (`.github/workflows/mobile.yml`)

- **Android**: syncs Capacitor, builds a debug APK, uploads it as a workflow artifact.
- **iOS**: builds an unsigned simulator `.app` on macOS and uploads a zip artifact.
  A device-installable IPA requires an Apple Developer signing identity; add a signing step before distribution.

Download the artifact from the workflow run, unzip, and install the APK on Android via
`adb install -r j3-debug.apk`. iOS simulator builds can be dragged onto the iOS Simulator.

## Releases (`.github/workflows/release.yml`)

Triggered by pushing a `v*` tag (or manual dispatch). It re-runs CI, builds mobile artifacts,
packages the Go binaries (Linux, macOS, Windows), copies the static web build, and publishes
everything as a GitHub release with generated notes.

## Local rigour

```sh
git config core.hooksPath .githooks   # enable pre-push checks
bun run test                          # Go + web unit tests
bun run check                         # types
bun run build                         # production build
bun run smoke                         # full-stack verification
bun run demo                          # manual QA
```

The pre-push hook rejects badly named branches and runs the test suite before pushing.
