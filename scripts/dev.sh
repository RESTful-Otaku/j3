#!/usr/bin/env bash
# J3 developer / operator console — nested TUI with keyboard navigation.
set -uo pipefail
export PATH="$HOME/.bun/bin:$PATH"
export J3_DEMO=1
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
  echo "  │   keyboard-navigable developer console       │"
  echo "  ╰──────────────────────────────────────────────╯"
  printf "%b\n" "${NC}"
  echo ""
}

select_menu() {
  local title="$1"; shift
  local -a items=("$@")
  local selected=0 key rest
  while true; do
    header
    printf "  %b%s%b\n" "$BOLD" "$title" "$NC"
    echo ""
    for i in "${!items[@]}"; do
      if [ "$i" -eq "$selected" ]; then
        printf "  %b➜  %s%b\n" "$GREEN" "${items[$i]}" "$NC"
      else
        printf "     %s\n" "${items[$i]}"
      fi
    done
    echo ""
    printf "  %b↑/↓ move  ·  Enter select  ·  Esc / q / b back%b\n" "$YELLOW" "$NC"
    read -rsn1 key
    case "$key" in
      $'\033')
        read -rsn2 -t 0.1 rest || rest=""
        case "$rest" in
          '[A') selected=$(( (selected - 1 + ${#items[@]}) % ${#items[@]} )) ;;
          '[B') selected=$(( (selected + 1) % ${#items[@]} )) ;;
          *) return 255 ;;
        esac
        ;;
      '') return "$selected" ;;
      [qQbB]) return 255 ;;
    esac
  done
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

pause() {
  echo ""
  read -r -p "  Press Enter to continue…" _
}

run_demo() {
  mkdir -p data
  spinner "Installing web dependencies" bun install --cwd web
  (
    trap 'kill %1 2>/dev/null || true' EXIT INT TERM
    go run ./cmd/j3d -addr 127.0.0.1:8787 -db ./data/j3-demo.db -seed &
    for _ in $(seq 1 80); do
      curl -fsS http://127.0.0.1:8787/healthz >/dev/null 2>&1 && break
      sleep 0.25
    done
    echo ""
    echo "  ✅ J3 demo ready → http://localhost:5173"
    echo ""
    bun run --cwd web dev
  )
}

run_demo_full() {
  mkdir -p data
  spinner "Installing web dependencies" bun install --cwd web
  spinner "Seeding complete tournament" go run ./cmd/j3d reset -full -db ./data/j3-demo.db
  (
    trap 'kill %1 2>/dev/null || true' EXIT INT TERM
    go run ./cmd/j3d -addr 127.0.0.1:8787 -db ./data/j3-demo.db -seed &
    for _ in $(seq 1 80); do
      curl -fsS http://127.0.0.1:8787/healthz >/dev/null 2>&1 && break
      sleep 0.25
    done
    echo ""
    echo "  ✅ Fully seeded tournament ready → http://localhost:5173"
    echo "     Brackets, tables, kata, medals and audit are all populated."
    echo ""
    bun run --cwd web dev
  )
}

qa_all() {
  spinner "Go tests" go test ./...
  spinner "Web type check" bun run check
  spinner "Web unit tests" bun run --cwd web test
  spinner "Full-stack smoke" bun run smoke
  spinner "Browser E2E + accessibility" bun run e2e
  spinner "Load test" bash scripts/load.sh
  spinner "Production build" bun run build
}

run_demo_phone() {
  mkdir -p data
  spinner "Installing web dependencies" bun install --cwd web
  echo ""
  echo "  📡 Venue server for phone QA: $(bash scripts/mobile.sh url)"
  echo "  ✅ Binds all interfaces with demo auth and CORS enabled."
  echo ""
  J3_DEMO=1 J3_ADDR=0.0.0.0:8787 J3_CORS_ORIGIN='*' \
    go run ./cmd/j3d -addr 0.0.0.0:8787 -db ./data/j3-demo.db -seed
}

open_api_docs() {
  (cd "$ROOT" && bun run sync:api-docs)
  local url="http://localhost:5173/apidocs"
  echo "  🔌 API explorer: $url"
  if command -v xdg-open >/dev/null 2>&1; then xdg-open "$url" >/dev/null 2>&1 & fi
}

qa_menu() {
  while true; do
    select_menu "QA & verification" \
      "Run all QA checks" "Full-stack smoke (52 checks)" "Full-data smoke (seeded tournament)" \
      "Browser E2E + axe accessibility" "Load test (200 concurrent)" "← Back"
    choice=$?
    [ "$choice" -eq 255 ] && return
    case "$choice" in
      0) qa_all ;;
      1) spinner "Full-stack smoke" bun run smoke ;;
      2) spinner "Full-data smoke" bun run smoke:full ;;
      3) spinner "Browser E2E" bun run e2e ;;
      4) spinner "Load test" bash scripts/load.sh ;;
      5) return ;;
    esac
    pause
  done
}

data_menu() {
  while true; do
    select_menu "Demo data & state" \
      "Reset demo data" "Seed demo data" "Seed full tournament" "Run live executable" "← Back"
    choice=$?
    [ "$choice" -eq 255 ] && return
    case "$choice" in
      0) spinner "Resetting demo data" go run ./cmd/j3d reset -db ./data/j3-demo.db ;;
      1) spinner "Seeding demo data" go run ./cmd/j3d seed -db ./data/j3-demo.db ;;
      2) spinner "Seeding full tournament" go run ./cmd/j3d reset -full -db ./data/j3-demo.db ;;
      3) J3_DB=./data/j3-live.db ./bin/j3d -addr 127.0.0.1:8787 ;;
      4) return ;;
    esac
    pause
  done
}

build_menu() {
  while true; do
    select_menu "Build & deploy" \
      "Build live executable + static UI" "Deploy to development" "Deploy to staging" \
      "Deploy to production" "← Back"
    choice=$?
    [ "$choice" -eq 255 ] && return
    case "$choice" in
      0) spinner "Building live executable" go build -o bin/j3d ./cmd/j3d; spinner "Building web app" bun run --cwd web build ;;
      1) spinner "Deploying development" bash scripts/deploy.sh development artifacts ;;
      2) spinner "Deploying staging" bash scripts/deploy.sh staging artifacts ;;
      3) spinner "Deploying production" bash scripts/deploy.sh production artifacts ;;
      4) return ;;
    esac
    pause
  done
}

mobile_menu() {
  while true; do
    select_menu "Mobile (Android / iOS)" \
      "Build demo APK" "Start Pixel emulator" "Install & launch APK" \
      "Build + install + launch" "Show phone server URL" "Run venue server for phone" \
      "Build iOS .app" "Install Playwright Chromium" "← Back"
    choice=$?
    [ "$choice" -eq 255 ] && return
    case "$choice" in
      0) bash scripts/mobile.sh build ;;
      1) bash scripts/mobile.sh emulator ;;
      2) bash scripts/mobile.sh install ;;
      3) bash scripts/mobile.sh build && bash scripts/mobile.sh install ;;
      4) echo "  📡 Phone URL: $(bash scripts/mobile.sh url)" ;;
      5) run_demo_phone ;;
      6) spinner "Building iOS app" bun run mobile:ios ;;
      7) spinner "Installing Chromium" bun run e2e:install ;;
      8) return ;;

    esac
    pause
  done
}

while true; do
  select_menu "J3 developer console" \
    "▶ Run full local demo" "🚀 Run fully seeded tournament" "🧪 QA & verification  ▸" \
    "💾 Demo data  ▸" "🏗 Build & deploy  ▸" "📱 Mobile  ▸" \
    "🔌 API explorer" "❌ Exit"
  choice=$?
  case "$choice" in
    0) run_demo ;;
    1) run_demo_full ;;
    2) qa_menu ;;
    3) data_menu ;;
    4) build_menu ;;
    5) mobile_menu ;;
    6) open_api_docs ;;
    7) echo "  👋 Bye!"; exit 0 ;;
  esac
done
