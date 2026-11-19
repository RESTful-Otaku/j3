#!/usr/bin/env bash
# J3 developer / operator console — friendly TUI for local QA, live builds and mobile.
set -uo pipefail

export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

GREEN='\033[0;32m'; CYAN='\033[0;36m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; BOLD='\033[1m'; NC='\033[0m'
SPIN=('⠋' '⠙' '⠹' '⠸' '⠼' '⠴' '⠦' '⠧' '⠇' '⠏')
LOG="/tmp/j3-task.log"

header() {
  clear 2>/dev/null || true
  printf "%b\n" "${CYAN}${BOLD}"
  echo "  ╭──────────────────────────────────────────────╮"
  echo "  │        J3 · TOURNAMENT CREATOR               │"
  echo "  │   dev / QA / live / mobile console           │"
  echo "  ╰──────────────────────────────────────────────╯"
  printf "%b\n" "${NC}"
}

spinner() {
  local msg="$1"; shift
  printf "  %b%s%b\n" "$YELLOW" "▶ $msg" "$NC"
  "$@" >"$LOG" 2>&1 &
  local pid=$! i=0
  while kill -0 "$pid" 2>/dev/null; do
    printf "\r    %b%s%b  %s" "$CYAN" "${SPIN[$i]}" "$NC" "$msg"
    i=$(( (i + 1) % ${#SPIN[@]} ))
    sleep 0.12
  done
  wait "$pid"; local code=$?
  if [ "$code" -eq 0 ]; then
    printf "\r    %b✅ %s%b                              \n" "$GREEN" "$msg" "$NC"
  else
    printf "\r    %b❌ %s (details: %s)%b\n" "$RED" "$msg" "$LOG" "$NC"
  fi
  return "$code"
}

run_demo() {
  mkdir -p data
  spinner "Installing web dependencies" bun install --cwd web
  printf "  %b🚀 Starting venue server + UI…%b\n" "$GREEN" "$NC"
  (
    trap 'kill %1 2>/dev/null || true' EXIT INT TERM
    go run ./cmd/j3d -addr 127.0.0.1:8787 -db ./data/j3-demo.db -seed &
    for _ in $(seq 1 80); do
      curl -fsS http://127.0.0.1:8787/healthz >/dev/null 2>&1 && break
      sleep 0.25
    done
    echo ""
    echo "  ✅ J3 demo ready"
    echo "     UI  → http://localhost:5173"
    echo "     API → http://127.0.0.1:8787/api/v1/demo"
    echo "     Press Ctrl-C to stop."
    echo ""
    bun run --cwd web dev
  )
}

qa_checks() {
  spinner "Go unit tests" go test ./...
  spinner "Svelte type check" bun run check
  spinner "Web tests" bun run --cwd web test
  spinner "Production build" bun run build
  echo "  ${GREEN}✅ All QA checks passed${NC}"
}

reset_data() { spinner "Resetting demo data" go run ./cmd/j3d reset -db ./data/j3-demo.db; }
seed_data()  { spinner "Seeding demo data"  go run ./cmd/j3d seed  -db ./data/j3-demo.db; }

build_live() {
  mkdir -p bin dist
  spinner "Building venue server binary" go build -o bin/j3d ./cmd/j3d
  spinner "Building static web app" bun run --cwd web build
  cp -R web/build/. dist/web/ 2>/dev/null || { mkdir -p dist/web && cp -R web/build/. dist/web/; }
  echo "  ${GREEN}✅ Live build ready:${NC} ./bin/j3d + ./dist/web"
}

run_live() {
  if [ ! -x bin/j3d ]; then echo "  ${RED}Build first (option 5)${NC}"; return; fi
  echo "  ${GREEN}▶ Running live executable on :8787 (Ctrl-C to stop)${NC}"
  J3_DB=./data/j3-live.db ./bin/j3d -addr 127.0.0.1:8787
}

mobile_sync() {
  spinner "Preparing web build" bun run --cwd web build
  spinner "Syncing Capacitor platforms" bunx --cwd web cap sync
}

android_build() {
  spinner "Preparing web build" bun run --cwd web build
  if ! ls web/android >/dev/null 2>&1; then spinner "Adding Android platform" bunx --cwd web cap add android; fi
  spinner "Syncing Capacitor" bunx --cwd web cap sync android
  spinner "Building Android APK (needs Android SDK)" bunx --cwd web cap build android
}

ios_build() {
  spinner "Preparing web build" bun run --cwd web build
  if ! ls web/ios >/dev/null 2>&1; then spinner "Adding iOS platform" bunx --cwd web cap add ios; fi
  spinner "Syncing Capacitor" bunx --cwd web cap sync ios
  if [ "$(uname)" != "Darwin" ]; then
    echo "  ${YELLOW}⚠ iOS builds require macOS + Xcode. Open web/ios/App/App.xcworkspace on a Mac.${NC}"
    return 1
  fi
  spinner "Building iOS app" bunx --cwd web cap build ios
}

while true; do
  header
  echo "  ${BOLD}Choose an option${NC}"
  echo "    1) 🚀 Run full local demo (server + UI, seeded)"
  echo "    2) 🧪 Run all QA checks"
  echo "    3) 🔄 Reset demo data"
  echo "    4) 🌱 Seed demo data"
  echo "    5) 🏗️  Build live executable + static UI"
  echo "    6) ▶️  Run live executable"
  echo "    7) 📱 Sync Capacitor platforms"
  echo "    8) 🤖 Build Android APK"
  echo "    9) 🍎 Build iOS app"
  echo "    0) 👋 Exit"
  echo ""
  read -r -p "  Choice: " choice
  case "$choice" in
    1) run_demo ;;
    2) qa_checks ;;
    3) reset_data ;;
    4) seed_data ;;
    5) build_live ;;
    6) run_live ;;
    7) mobile_sync ;;
    8) android_build ;;
    9) ios_build ;;
    0) echo "  👋 Bye!"; exit 0 ;;
    *) echo "  ${RED}Unknown option: $choice${NC}" ;;
  esac
  echo ""
  read -r -p "  Press Enter to continue…" _
done
