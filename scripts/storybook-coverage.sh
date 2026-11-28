#!/usr/bin/env bash
# Storybook coverage helper. Extensive, not a literal 100% guarantee without reviewing output.
set -euo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"; cd "$ROOT/web"
bun run build-storybook
echo "Stories:"
find src/stories -name '*.stories.svelte' | sort
echo "Built Storybook-static."
