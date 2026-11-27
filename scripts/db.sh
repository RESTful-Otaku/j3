#!/usr/bin/env bash
# Clean / purge / reseed / rebuild J3 SQLite for localised testing.
set -euo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

DB="${J3_DB:-${J3_DB_PATH:-./data/j3-demo.db}}"
mkdir -p "$(dirname "$DB")"

purge_db() {
  rm -f "$DB" "$DB-wal" "$DB-shm"
}

backup_db() {
  if [ -f "$DB" ]; then
    local ts; ts="$(date -u +%Y%m%dT%H%M%SZ)"
    cp "$DB" "$DB.bak.$ts"; echo "backup: $DB.bak.$ts"
  fi
}

mode="${1:-rebuild}"
case "$mode" in
  purge|clean|fresh)
    purge_db; echo "purged $DB" ;;
  reseed|fresh-reseed)
    purge_db
    go run ./cmd/j3d reset -okami -db "$DB"
    echo "fresh reseeded $DB" ;;
  seed)
    go run ./cmd/j3d seed -okami -db "$DB"
    echo "seeded $DB" ;;
  reset)
    go run ./cmd/j3d reset -okami -db "$DB"
    echo "reset $DB" ;;
  rebuild|fresh-full|full)
    purge_db
    go run ./cmd/j3d reset -okami -full -db "$DB"
    echo "rebuilt full $DB" ;;
  full-seed)
    go run ./cmd/j3d seed -okami -full -db "$DB"
    echo "full seed $DB" ;;
  existing|keep)
    echo "keeping existing $DB (no changes)" ;;
  backup)
    backup_db ;;
  *)
    echo "Usage: scripts/db.sh {purge|clean|fresh|reseed|fresh-reseed|seed|reset|rebuild|fresh-full|full|full-seed|existing|keep|backup}" >&2
    exit 1 ;;
esac
