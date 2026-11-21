#!/usr/bin/env bash
# Non-interactive full local demo: venue server + UI with seeded data.
set -euo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
mkdir -p data
bun install --cwd web >/dev/null
if [ "${J3_SEED_FULL:-0}" = "1" ]; then
  go run ./cmd/j3d reset -okami=false -full -db ./data/j3-demo.db
else
  go run ./cmd/j3d reset -okami -db ./data/j3-demo.db
fi
trap 'kill %1 2>/dev/null || true' EXIT INT TERM
go run ./cmd/j3d -addr 127.0.0.1:8787 -db ./data/j3-demo.db -seed &
for _ in $(seq 1 80); do
  curl -fsS http://127.0.0.1:8787/healthz >/dev/null 2>&1 && break
  sleep 0.25
done
echo "🚀 J3 demo ready → http://localhost:5173 (API: http://127.0.0.1:8787/api/v1/demo)"
bun run --cwd web dev
