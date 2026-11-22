#!/usr/bin/env bash
# Produce the offline demo assets bundled into the Android/iOS app:
#   * web/static/j3-demo.db        the actual seeded mid-event SQLite database
#   * web/static/j3-demo-data.json read-only API snapshot for phone screens
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

PORT="${J3_OFFLINE_PORT:-8799}"
DB="artifacts/j3-offline.db"
OUT="web/static/j3-demo-data.json"
SERVER_LOG="${TMPDIR:-/tmp}/j3-offline-server.log"
BIN="${TMPDIR:-/tmp}/j3-offline-bin"

mkdir -p artifacts web/static
rm -f "$DB" "${DB}-shm" "${DB}-wal" "$OUT" "$SERVER_LOG" "$BIN"

echo "🌱 Seeding bundled Okami 2027 mid-event database…"
go run ./cmd/j3d reset -okami -db "$DB" >/dev/null

echo "🚀 Starting temporary venue server on :$PORT…"
go build -o "$BIN" ./cmd/j3d
J3_DEMO=1 "$BIN" -addr "127.0.0.1:$PORT" -db "$DB" -seed >"$SERVER_LOG" 2>&1 &
SERVER_PID=$!
cleanup() {
  kill "$SERVER_PID" >/dev/null 2>&1 || true
  pkill -f "$BIN -addr 127.0.0.1:$PORT" >/dev/null 2>&1 || true
  rm -f "$BIN"
}
trap cleanup EXIT

for _ in $(seq 1 60); do
  if curl -fsS "http://127.0.0.1:$PORT/healthz" >/dev/null 2>&1; then break; fi
  sleep 1
done
curl -fsS "http://127.0.0.1:$PORT/healthz" >/dev/null || {
  echo "❌ Offline seed server did not start"; cat "$SERVER_LOG"; exit 1;
}

SESSION="$(curl -fsS -X POST "http://127.0.0.1:$PORT/api/v1/auth/demo")"
TOKEN="$(jq -r '.token' <<<"$SESSION")"
USER_ID="$(jq -r '.userId // ""' <<<"$SESSION")"
ROLE="$(jq -r '.role // "admin"' <<<"$SESSION")"
EMAIL="$(jq -r '.email // "j3@local"' <<<"$SESSION")"

ENDPOINTS=(
  /healthz
  /api/v1/version
  /api/v1/roles
  /api/v1/demo
  /api/v1/demo/accounts
  /api/v1/demo/credentials
  /api/v1/demo/attendance
  /api/v1/demo/belts
  /api/v1/demo/brackets
  /api/v1/demo/event
  "/api/v1/demo/event-config?eventId=ev-okami"
  /api/v1/demo/event-config
  /api/v1/demo/fights
  /api/v1/demo/identities
  /api/v1/demo/kata
  /api/v1/demo/kata/catalogue
  /api/v1/demo/medals
  /api/v1/demo/my-contests
  /api/v1/demo/announcements
  /api/v1/demo/officials
  /api/v1/demo/pools
  /api/v1/demo/query?table=clubs\&limit=100
  /api/v1/demo/readiness
  /api/v1/demo/role-requests
  /api/v1/demo/schedule
  "/api/v1/demo/simulate?seed=42&fighters=50"
  "/api/v1/demo/stats?seed=42&fighters=50"
  /api/v1/demo/standings
  /api/v1/demo/suggestions
  /api/v1/demo/team-rankings
  /api/v1/demo/training
  /api/v1/demo/view
  /api/v1/templates
)

TMP="$(mktemp)"
jq -n '{}' >"$TMP"
for path in "${ENDPOINTS[@]}"; do
  response="$(curl -fsS -H "Authorization: Bearer $TOKEN" "http://127.0.0.1:$PORT$path" || true)"
  if [[ -z "$response" ]] || ! jq -e . >/dev/null 2>&1 <<<"$response"; then
    echo "⚠️  Skipping $path"
    continue
  fi
  jq --arg key "$path" --argjson value "$response" '.[$key] = $value' "$TMP" >"${TMP}.next"
  mv "${TMP}.next" "$TMP"
done

jq \
  --arg userId "$USER_ID" \
  --arg role "$ROLE" \
  --arg email "$EMAIL" \
  --arg generated "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  '.[ "session" ] = {
      token: "offline-demo",
      email: $email,
      role: $role,
      expiresAt: "2099-01-01T00:00:00Z",
      userId: $userId
    }
    | .[ "generatedAt" ] = $generated
    | .[ "/healthz" ] = { status: "ok", offline: true }' \
  "$TMP" >"$OUT"
rm -f "$TMP"

cp "$DB" web/static/j3-demo.db
echo "✅ Bundled $(du -h "$OUT" | cut -f1) API snapshot at $OUT"
echo "✅ Bundled $(du -h web/static/j3-demo.db | cut -f1) SQLite database at web/static/j3-demo.db"
