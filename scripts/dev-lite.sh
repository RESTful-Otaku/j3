#!/usr/bin/env bash
# Pared-down developer TUI: only what matters for QA, config, build/run.
set -euo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

while true; do
  clear 2>/dev/null || true
  echo "  J3 Dev (lite)"
  echo "  1) Run local (dev)"
  echo "  2) Fresh DB (purge)"
  echo "  3) Reseed fresh"
  echo "  4) Rebuild full"
  echo "  5) Keep existing DB"
  echo "  6) QA gauntlet (checks+e2e+smoke+component)"
  echo "  7) Extended stress simulation"
  echo "  8) Build web flavors"
  echo "  9) Build QA mobile (APK)"
  echo "  10) Release all platforms"
  echo "  q) Exit"
  read -rp "  Choose: " c
  case "$c" in
    1) go run ./cmd/j3d -addr 127.0.0.1:8787 -db ./data/j3-demo.db -seed ;;
    2) bash scripts/db.sh fresh ;;
    3) bash scripts/db.sh reseed ;;
    4) bash scripts/db.sh rebuild ;;
    5) bash scripts/db.sh existing ;;
    6) bun run check && bun run test && bun run smoke && bash scripts/component-qa.sh ;;
    7) bash scripts/stress.sh ;;
    8) bun run build:flavors ;;
    9) bun run build:android:qa ;;
    10) bun run release ;;
    q|Q) exit 0 ;;
  esac
done
