#!/usr/bin/env bash
# Drive every TUI menu option with keystrokes and prove each action is dispatched.
set -uo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

LOG="$(mktemp /tmp/j3-tui-actions.XXXXXX)"
OUT="$(mktemp /tmp/j3-tui-output.XXXXXX)"
KEYS="$(mktemp /tmp/j3-tui-keys.XXXXXX)"
trap 'rm -f "$LOG" "$OUT" "$KEYS"' EXIT

python3 - "$KEYS" <<'PY'
import sys
seq = ''
def action(n):
    global seq
    seq += str(n) + '\n'          # numeric shortcut selects item n (1-based)
def back():
    global seq
    seq += 'q'                    # leave submenu

# Main: run demo, full demo, then every submenu.
action(1); action(2)
for menu in [3, 4, 5, 6, 7, 8]:
    action(menu)
    if menu == 3:  count = 7      # QA options + back
    elif menu == 4: count = 5     # data
    elif menu == 5: count = 9     # build
    elif menu == 6: count = 9     # mobile
    elif menu == 7: count = 9     # Docker QA + doctor/destroy; q backs
    else: count = 9              # docs & guides
    for i in range(1, count + 1):
        action(i)
    back()
action(9)                         # Exit
open(sys.argv[1], 'w').write(seq)
PY

J3_TUI_TEST=1 J3_TUI_LOG="$LOG" bash scripts/dev.sh < "$KEYS" > "$OUT" 2>&1

expected="run-demo run-demo-full qa-menu qa-all qa-smoke qa-smoke-full qa-e2e qa-load qa-github qa-back \
data-menu data-reset data-seed data-seed-full data-run-live menu-back \
build-menu build-exec build-release build-cicd deploy-dev deploy-staging deploy-prod \
build-and-deploy-staging build-and-deploy-production menu-back \
mobile-menu mobile-build-apk mobile-emulator mobile-install mobile-build-install mobile-url \
run-demo-phone mobile-build-ios mobile-build-hosted mobile-back \
docker-menu docker-full docker-okami docker-clean docker-light docker-dev docker-build docker-down docker-reset docker-doctor \
docs-menu docs-home docs-api docs-guide docs-development docs-ci docs-hosting docs-guides \
docs-refresh docs-back \
exit"

missing=0
for id in $expected; do
  if ! grep -qx "$id" "$LOG"; then
    echo "❌ TUI action not dispatched: $id"
    missing=$((missing + 1))
  fi
done

if [ "$missing" -ne 0 ]; then
  echo "❌ TUI harness failed: $missing actions missing"
  tail -40 "$OUT"
  exit 1
fi

if grep -qE 'Unknown option|command not found|syntax error' "$OUT"; then
  echo "❌ TUI emitted errors"
  grep -E 'Unknown option|command not found|syntax error' "$OUT"
  exit 1
fi

echo "✅ TUI harness passed: $(echo $expected | wc -w) actions dispatched"
