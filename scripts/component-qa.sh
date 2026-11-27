#!/usr/bin/env bash
# Svelte Storybook-equivalent component QA on /ui.
set -euo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/web"
bunx playwright test component-gallery.spec.ts --reporter=list
