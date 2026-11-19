#!/usr/bin/env bash
# One command to run the whole J3 app locally with seeded demo data.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! command -v bun >/dev/null 2>&1; then
  export PATH="$HOME/.bun/bin:$PATH"
fi

mkdir -p data
bun install --cwd web >/dev/null

trap 'kill %1 2>/dev/null || true' EXIT INT TERM

go run ./cmd/j3d -addr 127.0.0.1:8787 -db ./data/j3-demo.db -seed &

for _ in $(seq 1 80); do
  if curl -fsS http://127.0.0.1:8787/healthz >/dev/null 2>&1; then
    break
  fi
  sleep 0.25
done

echo ""
echo "  J3 demo ready"
echo "  UI  → http://localhost:5173"
echo "  API → http://127.0.0.1:8787/api/v1/demo"
echo "  Reset demo data with: bun run reset"
echo ""

bun run --cwd web dev
