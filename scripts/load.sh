#!/usr/bin/env bash
# Lightweight load / stability check against a running venue server.
set -uo pipefail
BASE="${LOAD_BASE:-http://127.0.0.1:8787}"
REQUESTS="${LOAD_REQUESTS:-200}"
PARALLEL="${LOAD_PARALLEL:-20}"

echo "🚀 Load test: $REQUESTS requests, $PARALLEL concurrent → $BASE"
failures=0

check_url() {
  local code
  code="$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 "$1")"
  case "$code" in
    200|201) ;;
    *) echo "❌ $1 → $code"; echo 1 ;;
  esac
}

export -f check_url
export BASE

seq 1 "$REQUESTS" | xargs -P "$PARALLEL" -I{} bash -c '
  case $(( {} % 4 )) in
    0) path="/healthz" ;;
    1) path="/api/v1/demo" ;;
    2) path="/api/v1/demo/fights" ;;
    3) path="/api/v1/demo/standings" ;;
  esac
  check_url "$BASE$path"
' | grep -c '^1$' > /tmp/j3-load-failures 2>/dev/null || true

failures=$(cat /tmp/j3-load-failures 2>/dev/null || echo 0)
if [ "$failures" -ne 0 ]; then
  echo "❌ $failures failed requests"
  exit 1
fi
echo "✅ All $REQUESTS requests succeeded"
