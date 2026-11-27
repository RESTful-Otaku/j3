# Testing & release-grade gauntlets

`svelte-check`; `go test ./... -race`; `go vet`; `bun run --cwd web test`; `bun run smoke`/`smoke:full`; `bun run e2e` (Playwright + axe, now incl component gallery + mobile/tablet + safe-area + 320px); `bash scripts/component-qa.sh` (Svelte Storybook-equivalent `/ui`); `scripts/load.sh`; `check:workflows`; generated-type parity.

**Extended deterministic simulation** `j3d stress`, `RunExtended`, `bun run simulate:extended`. `scripts/stress.sh` `bun run stress` full gauntlet. Deterministic multi-user, multi-role, multi-mutation, compressed hours, `verifyOK`, periodic snapshots. Re-run same seed diff.

**DB** see docs/DB.md. **Builds** `build:flavors`, `build:android:*`, `build-ios`, `release`. **TUI** `bun run dev`, `bun run dev:lite`. **CI/CD** `bun run cicd`, hosted CI/Mobile/Release/Deploy green; free-only public; visibility-gated.
