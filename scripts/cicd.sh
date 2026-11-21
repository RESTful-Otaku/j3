#!/usr/bin/env bash
# Local equivalent of the GitHub Actions CI gates.
set -euo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

bun run check:workflows
# Load suite needs a live venue server; start one for the local CI run.
go build -o /tmp/j3-cicd-bin ./cmd/j3d
/tmp/j3-cicd-bin -addr 127.0.0.1:8787 -db /tmp/j3-cicd.db -seed > /tmp/j3-cicd-server.log 2>&1 &
SERVER_PID=$!
trap 'kill "$SERVER_PID" 2>/dev/null || true' EXIT
for _ in $(seq 1 90); do
  curl -fsS http://127.0.0.1:8787/healthz >/dev/null 2>&1 && break
  sleep 1
done
go vet ./...
go test ./...
go test -race ./...
bun run check
bun run test
bun run smoke:full
bun run load
bun run e2e
bun run build
echo "✅ All CI/CD checks passed locally"
