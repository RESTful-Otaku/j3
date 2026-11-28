#!/usr/bin/env bash
# Fully reset app DB + local QA artifacts to a fresh seeded local demo (web + mobile).
set -euo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"; cd "$ROOT"
TARGET="${1:-local}" # local|web|mobile|docker|all
DB="${J3_DB:-${J3_DB_PATH:-./data/j3-demo.db}}"

log() { printf '\n==> %s\n' "$*"; }

reset_db() {
  log "Purge SQLite + wal/shm + reseed Okami + full"
  rm -f "$DB" "$DB-wal" "$DB-shm" "${DB%.db}.db-wal" "${DB%.db}.db-shm" "$DB.bak."* 2>/dev/null || true
  go run ./cmd/j3d reset -okami -full -db "$DB"
}

reset_web_static() {
  log "Regenerate bundled web demo data"
  bash scripts/mobile-demo-data.sh
  cp web/static/j3-demo.db web/static/j3-demo-data.json ./ 2>/dev/null || true
  # restore tracked generated source after copy attempt
  git checkout -- web/static/j3-demo.db web/static/j3-demo-data.json 2>/dev/null || true
}

case "$TARGET" in
  local|web)
    reset_db
    reset_web_static
    ;;
  mobile)
    reset_db
    bash scripts/mobile-demo-data.sh
    bun run build:android:qa
    ;;
  docker)
    J3_CONTAINER_RUNTIME="${J3_CONTAINER_RUNTIME:-}" bash scripts/qa-docker.sh destroy || true
    J3_CONTAINER_RUNTIME="${J3_CONTAINER_RUNTIME:-}" bash scripts/qa-docker.sh full
    ;;
  all)
    reset_db; reset_web_static; bun run build:android:qa
    ;;
  *) echo "usage: scripts/reset-app.sh [local|web|mobile|docker|all]" >&2; exit 1;;
esac

log "Reset complete. In web/mobile UI run Settings → Reset app & data to clear localStorage/caches."
