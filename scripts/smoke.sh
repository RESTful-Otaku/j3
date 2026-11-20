#!/usr/bin/env bash
# Automated end-to-end smoke test: boots the real venue server and verifies
# every critical endpoint and UI route. Exits non-zero on any failure.
set -uo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

PORT="${SMOKE_PORT:-8799}"
BASE="http://127.0.0.1:$PORT"
DB="/tmp/j3-smoke-$$.db"
PASS=0
FAIL=0

cleanup() {
  kill "$SERVER_PID" 2>/dev/null || true
  rm -f "$DB" "$DB-wal" "$DB-shm"
}
trap cleanup EXIT INT TERM

check() {
  local label="$1" expected="$2" actual="$3"
  if [ "$actual" = "$expected" ]; then
    printf "  ✅ %-42s %s\n" "$label" "$actual"
    PASS=$((PASS + 1))
  else
    printf "  ❌ %-42s got %s, want %s\n" "$label" "$actual" "$expected"
    FAIL=$((FAIL + 1))
  fi
}

status() { curl -s -o /dev/null -w '%{http_code}' "$@"; }
has() { curl -s "$1" | grep -q "$2" && echo 1 || echo 0; }

echo "🚀 Starting venue server on $PORT"
go run ./cmd/j3d -addr "127.0.0.1:$PORT" -db "$DB" -seed >/tmp/j3-smoke-server.log 2>&1 &
SERVER_PID=$!
for _ in $(seq 1 80); do
  curl -fsS "$BASE/healthz" >/dev/null 2>&1 && break
  sleep 0.25
done

echo "🔎 API checks"
check "healthz" "200" "$(status "$BASE/healthz")"
check "readyz" "200" "$(status "$BASE/readyz")"
check "roles" "200" "$(status "$BASE/api/v1/roles")"
check "demo counts" "200" "$(status "$BASE/api/v1/demo")"
check "demonstration accounts" "200" "$(status "$BASE/api/v1/demo/accounts")"
check "fights" "200" "$(status "$BASE/api/v1/demo/fights")"
check "standings" "200" "$(status "$BASE/api/v1/demo/standings")"
check "event details" "200" "$(status "$BASE/api/v1/demo/event")"
check "belts" "200" "$(status "$BASE/api/v1/demo/belts")"
check "schedule" "200" "$(status "$BASE/api/v1/demo/schedule")"
check "kata" "200" "$(status "$BASE/api/v1/demo/kata")"
check "medals" "200" "$(status "$BASE/api/v1/demo/medals")"
check "kata catalogue" "200" "$(status "$BASE/api/v1/demo/kata/catalogue")"
check "training sessions" "200" "$(status "$BASE/api/v1/demo/training")"
check "audit" "200" "$(status "$BASE/api/v1/demo/audit")"
check "query" "200" "$(status "$BASE/api/v1/query?table=fighters&limit=5")"
check "simulation" "200" "$(status "$BASE/api/v1/demo/simulate?fighters=50")"
check "stats" "200" "$(status "$BASE/api/v1/demo/stats?fighters=50")"

echo "🔎 Workflow checks"
check "sign-up preview" "200" "$(status -X POST "$BASE/api/v1/signup/preview" -H 'Content-Type: application/json' -d '{"dateOfBirth":"1998-01-01","unitSystem":"metric","heightCm":175,"weightKg":74,"gender":"male"}')"
check "bracket build" "201" "$(status -X POST "$BASE/api/v1/demo/brackets/build" -H 'Content-Type: application/json' -d '{"eventId":"ev-demo","format":"Double elimination","targetPoolSize":4,"seed":7,"thirdPlace":true}')"
check "pool generation" "200" "$(status -X POST "$BASE/api/v1/demo/pools/generate" -H 'Content-Type: application/json' -d '{"eventId":"ev-demo","categoryId":"cat-adult-76","targetPoolSize":4,"seed":9}')"
check "fight start" "200" "$(status -X POST "$BASE/api/v1/fights/fight-01/start" -H 'Content-Type: application/json' -d '{"seconds":120,"round":1}')"
check "fight score" "200" "$(status -X POST "$BASE/api/v1/fights/fight-01/event" -H 'Content-Type: application/json' -d '{"corner":"red","points":2,"technique":"Takedown","kind":"score"}')"
check "fight pause" "200" "$(status -X POST "$BASE/api/v1/fights/fight-01/pause" -H 'Content-Type: application/json' -d '{}')"
check "fight hold" "200" "$(status -X POST "$BASE/api/v1/fights/fight-01/hold" -H 'Content-Type: application/json' -d '{"corner":"red","seconds":20}')"
check "assign official" "200" "$(status -X POST "$BASE/api/v1/demo/fights/fight-01/officials" -H 'Content-Type: application/json' -d '{"userId":"u-official","role":"referee"}')"
check "fight submit" "200" "$(status -X POST "$BASE/api/v1/fights/fight-01/submit" -H 'Content-Type: application/json' -d '{"outcome":"win_red"}')"
check "kata round create" "201" "$(status -X POST "$BASE/api/v1/demo/kata/performances" -H 'Content-Type: application/json' -d '{"fighterId":"f07","kataName":"Hekigan-ken","roundNumber":2,"kataType":"shitei","panelSize":5}')"
check "schedule drift" "200" "$(status -X POST "$BASE/api/v1/demo/schedule/drift" -H 'Content-Type: application/json' -d '{"minutes":5}')"
check "medal award" "200" "$(status -X POST "$BASE/api/v1/demo/medals/award" -H 'Content-Type: application/json' -d '{"gold":"f01","silver":"f02","bronze":["f03","f04"]}')"
check "tie-break standings" "200" "$(status "$BASE/api/v1/demo/standings?chain=wins,submissions,points_scored")"
check "sparring rotation" "200" "$(status -X POST "$BASE/api/v1/demo/rotation" -H 'Content-Type: application/json' -d '{"fighters":["f01","f02","f03","f04"],"mats":["Mat 1","Mat 2"],"rounds":3}')"

echo "🔎 Data assertions"
check "simulator has medals" "1" "$(has "$BASE/api/v1/demo/simulate?fighters=50" '"medals"')"
check "query rejects bad table" "400" "$(status "$BASE/api/v1/query?table=sqlite_master")"
check "signup rejects bad email" "400" "$(status -X POST "$BASE/api/v1/signup" -H 'Content-Type: application/json' -d '{"name":"A","email":"bad","dateOfBirth":"1990-01-01"}')"
check "fighter cannot submit result" "403" "$(status -X POST "$BASE/api/v1/fights/fight-01/submit" -H 'Content-Type: application/json' -H 'X-J3-Role: fighter' -d '{"outcome":"win_red"}')"

echo "🔎 UI routes"
export PATH="$HOME/.bun/bin:$PATH"
if command -v bun >/dev/null 2>&1; then
  (cd web && bun run dev >/tmp/j3-smoke-web.log 2>&1) &
  WEB_PID=$!
  for _ in $(seq 1 60); do
    curl -fsS http://127.0.0.1:5173/ >/dev/null 2>&1 && break
    sleep 0.25
  done
  for route in / /setup /display /mat /public /brackets /tables /pools /control /data /audit /settings /details /simulate /ceremony; do
    check "UI $route" "200" "$(status "http://127.0.0.1:5173$route")"
  done
  kill "$WEB_PID" 2>/dev/null || true
fi

echo ""
echo "  Passed: $PASS  Failed: $FAIL"
[ "$FAIL" -eq 0 ] || exit 1
