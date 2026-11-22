#!/usr/bin/env bash
# Launch the J3 documentation site (guides, developer docs and API reference).
# Starts the local venue server and UI only if they are not already running.
set -uo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

PAGE="${1:-/docs}"
case "$PAGE" in
  /docs|/docs/*) ;;
  *) echo "❌ Only /docs pages can be opened by this helper (got: $PAGE)" >&2; exit 2 ;;
esac

echo "🔄 Refreshing guides and API index…"
bun run sync:api-docs >/dev/null || exit 1
mkdir -p data

if ! curl -fsS http://127.0.0.1:8787/healthz >/dev/null 2>&1; then
  echo "▶ Starting local venue server…"
  J3_DEMO=1 J3_CORS_ORIGIN='http://localhost:5173,http://127.0.0.1:5173,http://localhost,capacitor://localhost' \
    nohup go run ./cmd/j3d -addr 127.0.0.1:8787 -db ./data/j3-demo.db -seed \
    >/tmp/j3-docs-server.log 2>&1 &
fi
if ! curl -fsS http://127.0.0.1:5173/ >/dev/null 2>&1; then
  echo "▶ Starting local UI…"
  nohup bun run --cwd web dev >/tmp/j3-docs-ui.log 2>&1 &
fi

ready=0
for _ in $(seq 1 120); do
  if curl -fsS http://127.0.0.1:8787/healthz >/dev/null 2>&1 &&
     curl -fsS -o /dev/null -w '%{content_type}' http://127.0.0.1:5173/docs | grep -q 'text/html'; then
    ready=1
    break
  fi
  sleep 0.5
done
if [ "$ready" -ne 1 ]; then
  echo "❌ Docs site did not start. Check /tmp/j3-docs-server.log and /tmp/j3-docs-ui.log" >&2
  exit 1
fi

URL="http://localhost:5173$PAGE"
echo "📚 J3 docs ready → $URL"
if [ -n "${BROWSER:-}" ]; then
  "$BROWSER" "$URL" >/dev/null 2>&1 &
elif command -v xdg-open >/dev/null 2>&1; then
  xdg-open "$URL" >/dev/null 2>&1 &
elif command -v open >/dev/null 2>&1; then
  open "$URL" >/dev/null 2>&1 &
fi
