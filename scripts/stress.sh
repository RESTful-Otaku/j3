#!/usr/bin/env bash
# Extended deterministic simulation + full gauntlet for production-grade QA.
set -euo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "=== 1. Extended deterministic multi-user simulation ==="
go run ./cmd/j3d stress -seed "${STRESS_SEED:-42}" -fighters "${STRESS_FIGHTERS:-50}" \
  -users "${STRESS_USERS:-8}" -steps "${STRESS_STEPS:-256}" -out build/stress-extended.json
python3 - <<'PYEOF' 2>/dev/null || true
import json,os
p='build/stress-extended.json'
if os.path.exists(p):
 d=json.load(open(p)); print('verifyOK=',d['verifyOK'],'users=',d['users'],'mutations=',d['mutations'],'snapshots=',len(d['snapshots']))
PYEOF

echo "=== 2. Fresh deterministic DB ==="
bash scripts/db.sh rebuild

echo "=== 3. Go tests (race) ==="
go test ./... -race

echo "=== 4. Go vet ==="
go vet ./...

echo "=== 5. svelte-check ==="
bun run check

echo "=== 6. Web unit tests ==="
bun run --cwd web test

echo "=== 7. Full smoke ==="
bun run smoke:full

echo "=== 8. Browser E2E + accessibility ==="
bun run e2e

echo "=== 9. Component QA (Storybook-equivalent gallery) ==="
bash scripts/component-qa.sh

echo "=== 10. Load ==="
bash scripts/load.sh

echo "=== 11. Workflow check ==="
bun run check:workflows

echo "=== 12. Build all web flavors + QA mobile ==="
bun run build:flavors
bun run build:android:qa

echo "✅ Gauntlet complete"
