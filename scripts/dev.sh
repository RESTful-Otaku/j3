#!/usr/bin/env bash
# J3 developer / operator console — nested TUI with keyboard navigation.
set -uo pipefail
export PATH="$HOME/.bun/bin:$PATH"
export J3_DEMO=1
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

GREEN='\033[0;32m'; CYAN='\033[0;36m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; BOLD='\033[1m'; NC='\033[0m'
ORANGE='\033[38;5;214m'; PINK='\033[38;5;205m'; DIM='\033[2m'; INV='\033[7m'; CLEARLINE='\033[K'
SPIN=('⠋' '⠙' '⠹' '⠸' '⠼' '⠴' '⠦' '⠧' '⠇' '⠏')
LOG="/tmp/j3-task.log"

header() {
  clear 2>/dev/null || true
  printf "%b\n" "${CYAN}${BOLD}"
  echo "  ╭────────────────────────────────────────────────╮"
  echo "  │  🥋  J3 · TOURNAMENT CREATOR                    │"
  echo "  │      Japanese Jiu Jitsu platform                │"
  echo "  │      keyboard-navigable developer console       │"
  echo "  ╰────────────────────────────────────────────────╯"
  echo "   ${DIM}J3${NC}${CYAN}${BOLD} ─ build · QA · run · verify · release${NC}"
  echo "   ${DIM}ASCII:  __  _____      __  →  J3  🥋${NC}"
  echo "   ${DIM}   | |/   |__/ __| ___ _ _ ___ ___ _ _  _| |_${NC}"
  echo "   ${DIM}  _|_|    |__\\__ \/ _ \ '_/ -_) _ \ ' \/ _|  _|${NC}"
  echo "   ${DIM} |____|      |___/\\___/_| \\___\\___/_||_\\__|\\__|${NC}"
  echo ""
}

menu_gap() { # ensure a space between leading emoji and label
  local item="$1"
  if printf '%s' "$item" | grep -q ' '; then printf '%s' "$item"; return; fi
  printf '%s' "$item" | sed -E 's/^([^[:space:]]+)([A-Za-z])/\1 \2/'
}

menu_desc() { # short description shown under each option
  case "$1" in
    *Run\ full\ local\ demo*) echo "Seeded local venue + web; no rebuild." ;;
    *Run\ fully\ seeded\ tournament*) echo "Okami mid-event full seed; brackets/tables/kata/medals/audit populated." ;;
    *QA\ \&\ verification*|*QA\ and\ verification*) echo "Checks, smoke, E2E+axe, load, component QA, CI." ;;
    *Run\ all\ QA\ checks*) echo "Go+web+smoke+E2E+load+build." ;;
    *Demo\ data*|*Demo\ data\ \&\ state*) echo "Fresh / reseeded / full / keep-existing DB." ;;
    *Build\ \&\ deploy*|*Build\ \&\ deploy\ \>*|*Build\ \&\ Deploy*) echo "Flavors, release, CI/CD, deploy dev/staging/prod." ;;
    *Build\ live\ executable*) echo "Go server + static UI into build/." ;;
    *Build\ release*) echo "Server+web+API+docs release." ;;
    *Run\ full\ CI/CD*) echo "Local full CI/CD suite." ;;
    *Deploy\ to\ development*) echo "deploy.sh development." ;;
    *Deploy\ to\ staging*) echo "deploy.sh staging." ;;
    *Deploy\ to\ production*) echo "deploy.sh production (confirm)." ;;
    *Build\ +\ deploy\ staging*) echo "Release then deploy staging." ;;
    *Build\ +\ deploy\ production*) echo "Release then deploy production." ;;
    *Mobile*) echo "Android/iOS build+install, emulator, hosted HTTPS." ;;
    *Build\ demo\ APK*) echo "Standalone offline QA APK." ;;
    *Start\ Pixel*) echo "Pixel AVD." ;;
    *Install\ \&\ launch*) echo "Install + launch on device." ;;
    *Build\ +\ install*) echo "Build then install+launch." ;;
    *Show\ phone\ server\ URL*) echo "LAN URL for phone." ;;
    *Run\ venue\ server\ for\ phone*) echo "0.0.0.0 + CORS + demo." ;;
    *Build\ iOS*) echo "iOS simulator app." ;;
    *Build\ APK\ for\ hosted*) echo "Prompts HTTPS API; J3_SERVER_URL." ;;
    *Docker\ QA*) echo "Container QA stack." ;;
    *Full\ seeded\ tournament*) echo "Docker full seed." ;;
    *Okami\ mid-event*) echo "Docker Okami." ;;
    *Clean\ base\ seed*) echo "Docker clean." ;;
    *Light\ 24-fighter*) echo "Docker light." ;;
    *Dev\ hot\ reload*) echo "Docker dev." ;;
    *Build\ container\ images*|*Build\ Container\ Images*|*Build\ container*) echo "Docker/Podman images." ;;
    *Stop\ containers*) echo "Docker down." ;;
    *Reset\ QA\ data*) echo "Docker reset." ;;
    *Docs*) echo "Docs/API/guides." ;;
    *Fresh\ DB*) echo "Purge SQLite." ;;
    *Reseed*) echo "Fresh + seeded." ;;
    *Rebuild\ full*) echo "Fresh full." ;;
    *Keep\ existing*) echo "No DB change." ;;
    *Exit*) echo "Quit." ;;
    *Back*) echo "Back." ;;
    *) echo "" ;;
  esac
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
      local label; label="$(menu_gap "${items[$i]}")"
      local desc; desc="$(menu_desc "${items[$i]}")"
      if [ "$i" -eq "$selected" ]; then
        printf "  %b➜  %s%b %s%s%b\n" "$GREEN$BOLD" "$label" "$NC" "$CYAN" "" "$NC"
        if [ -n "$desc" ]; then printf "      %b%s%b\n" "$DIM" "$desc" "$NC"; fi
      else
        printf "     %s\n" "$label"
        if [ -n "$desc" ]; then printf "      %b%s%b\n" "$DIM" "$desc" "$NC"; fi
      fi
    done
    echo ""
    printf "  %b↑/↓ move  ·  Enter select  ·  Esc / q / b back  ·  1-9 shortcut%b\n" "$YELLOW" "$NC"
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
      [1-9]) return $((10#$key - 1)) ;;
      [qQbB]) return 255 ;;
    esac
  done
}

tui_log() {
  if [ -n "${J3_TUI_LOG:-}" ]; then
    printf '%s\n' "$1" >> "$J3_TUI_LOG"
  fi
}

# In test mode, stub only actions that would block on a server or emulator.
test_stub() {
  local label="$1"
  if [ "${J3_TUI_TEST:-0}" = "1" ]; then
    echo "  ✅ $label (test mode)"
    return 0
  fi
  return 1
}

now_ms() { date +%s%3N 2>/dev/null || python3 -c 'import time;print(int(time.time()*1000))'; }
progress_bar() { local cur="${1:-0}" total="${2:-100}" width="${3:-12}"; local filled=$(( cur * width / total )); local bar=""; for ((i=0;i<filled;i++)); do bar+="█"; done; for ((i=filled;i<width;i++)); do bar+="░"; done; printf '%s' "$bar"; }
log_start() { printf "  %b▶ %s%b\n" "$YELLOW" "$1" "$NC"; }
log_info() { printf "  %bℹ %s%b\n" "$CYAN" "$1" "$NC"; }
log_ok() { printf "  %b✔ PASS %s%b\n" "$GREEN$BOLD" "$1" "$NC"; }
log_warn() { printf "  %b⚠ WARN %s%b\n" "$YELLOW$BOLD" "$1" "$NC"; }
log_fail() { printf "  %b✖ FAIL %s%b\n" "$RED$BOLD" "$1" "$NC"; }
log_section() { printf "\n%b▔▔ %s ▔▔%b\n" "$PINK$BOLD" "$1" "$NC"; }

spinner() {
  local msg="$1"; shift
  if [ "${J3_TUI_TEST:-0}" = "1" ]; then
    printf "  ✅ %s (test mode)\n" "$msg"
    return 0
  fi
  log_start "$msg"
  local start_ms; start_ms=$(now_ms)
  "$@" >"$LOG" 2>&1 &
  local pid=$! i=0
  while kill -0 "$pid" 2>/dev/null; do
    local bar; bar="$(progress_bar $(( (i % 10) * 10 )) 100 12)"
    printf "\r    %b%s%b %s %b%s%b%s" "$CYAN" "${SPIN[$i]}" "$NC" "$msg" "$DIM" "$bar" "$NC" "$CLEARLINE"
    i=$(( (i + 1) % ${#SPIN[@]} ))
    sleep 0.12
  done
  wait "$pid"; local code=$?
  local end_ms; end_ms=$(now_ms); local ms=$((end_ms - start_ms))
  if [ "$code" -eq 0 ]; then
    printf "\r    %b✔ PASS%b %s %b(%s ms)%b%s\n" "$GREEN$BOLD" "$NC" "$msg" "$DIM" "$ms" "$NC" "$CLEARLINE"
  else
    printf "\r    %b✖ FAIL%b %s %b(%s ms; log %s)%b%s\n" "$RED$BOLD" "$NC" "$msg" "$DIM" "$ms" "$LOG" "$NC" "$CLEARLINE"
    echo "    ${DIM}── log tail ──${NC}"; tail -n 15 "$LOG" | sed 's/^/    /'
  fi
  return "$code"
}

pause() {
  echo ""
  read -r -p "  Press Enter to continue…" _
}

run_demo() {
  if test_stub "run-demo"; then return; fi
  mkdir -p data
  spinner "Installing web dependencies" bun install --cwd web
  (
    trap 'kill %1 2>/dev/null || true' EXIT INT TERM
    J3_CORS_ORIGIN='http://localhost:5173,http://127.0.0.1:5173,http://localhost,capacitor://localhost' \
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
  if test_stub "run-demo-full"; then return; fi
  mkdir -p data
  spinner "Installing web dependencies" bun install --cwd web
  spinner "Seeding complete tournament" go run ./cmd/j3d reset -okami -full -db ./data/j3-demo.db
  (
    trap 'kill %1 2>/dev/null || true' EXIT INT TERM
    J3_CORS_ORIGIN='http://localhost:5173,http://127.0.0.1:5173,http://localhost,capacitor://localhost' \
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
  log_section "QA gauntlet"
  local started; started=$(now_ms)
  local pass=0 fail=0 warn=0
  run_qa() { local name="$1"; shift; if spinner "$name" "$@"; then pass=$((pass+1)); else fail=$((fail+1)); fi; }
  run_qa "Go tests" go test ./...
  run_qa "Go vet" go vet ./...
  run_qa "Web type check" bun run check
  run_qa "Web unit tests" bun run --cwd web test
  run_qa "Full-stack smoke" bun run smoke
  run_qa "Browser E2E + accessibility" bun run e2e
  run_qa "Component QA" bash scripts/component-qa.sh
  run_qa "Load test" bash scripts/load.sh
  run_qa "Workflow check" bun run check:workflows
  run_qa "Production build" bun run build
  local ms=$(( $(now_ms) - started ))
  log_section "Summary"
  printf "  %b✔ PASS %d%b  %b⚠ WARN %d%b  %b✖ FAIL %d%b  %b(%s ms)%b\n" "$GREEN$BOLD" "$pass" "$NC" "$YELLOW$BOLD" "$warn" "$NC" "$RED$BOLD" "$fail" "$NC" "$DIM" "$ms" "$NC"
  printf "  Logs: %s  ·  build/  ·  Storybook: bun run storybook%b\n" "$LOG" "$NC"
  [ "$fail" -eq 0 ]
}

run_demo_phone() {
  if test_stub "run-demo-phone"; then return; fi
  mkdir -p data
  spinner "Installing web dependencies" bun install --cwd web
  echo ""
  echo "  📡 Venue server for phone QA: $(bash scripts/mobile.sh url)"
  echo "  ✅ Binds all interfaces with demo auth and CORS enabled."
  echo ""
  J3_DEMO=1 J3_ADDR=0.0.0.0:8787 J3_CORS_ORIGIN='http://localhost,capacitor://localhost' \
    go run ./cmd/j3d -addr 0.0.0.0:8787 -db ./data/j3-demo.db -seed
}

open_docs() {
  local page="${1:-/docs}"
  if test_stub "open-docs"; then return; fi
  bash scripts/docs.sh "$page"
}

docs_menu() {
  while true; do
    select_menu "Docs & developer guides" \
      "📚 Open documentation site" "🔌 API reference (try endpoints)" \
      "🏁 Organiser guide" "🧑‍💻 Developer guide" "🚦 CI & releases" \
      "☁️ Free hosting plan" "🎓 Hands-on guides" \
      "🔄 Refresh generated API index" "← Back"
    choice=$?
    [ "$choice" -eq 255 ] && return
    case "$choice" in
      0) tui_log docs-home; open_docs /docs ;;
      1) tui_log docs-api; open_docs /docs/api ;;
      2) tui_log docs-guide; open_docs /docs/guide ;;
      3) tui_log docs-development; open_docs /docs/development ;;
      4) tui_log docs-ci; open_docs /docs/ci ;;
      5) tui_log docs-hosting; open_docs /docs/hosting ;;
      6) tui_log docs-guides; open_docs /guides ;;
      7) tui_log docs-refresh; spinner "Refreshing docs index" bun run sync:api-docs ;;
      8) tui_log docs-back; return ;;
    esac
    pause
  done
}

qa_menu() {
  while true; do
    select_menu "QA & verification" \
      "Run all QA checks" "Full-stack smoke" "Full-data smoke (seeded tournament)" \
      "Browser E2E + axe accessibility" "Load test (200 concurrent)" \
      "Verify GitHub Actions" "← Back"
    choice=$?
    [ "$choice" -eq 255 ] && return
    case "$choice" in
      0) tui_log qa-all; qa_all ;;
      1) tui_log qa-smoke; spinner "Full-stack smoke" bun run smoke:full ;;
      2) tui_log qa-smoke-full; spinner "Full-data smoke" bun run smoke:full ;;
      3) tui_log qa-e2e; spinner "Browser E2E" bun run e2e ;;
      4) tui_log qa-load; spinner "Load test" bash scripts/load.sh ;;
      5) tui_log qa-github; spinner "GitHub Actions" gh run list --repo RESTful-Otaku/j3 --limit 5 ;;
      6) tui_log qa-back; return ;;
    esac
    pause
  done
}

docker_menu() {
  while true; do
    select_menu "Docker QA demo" \
      "🐳 Full seeded tournament" "🥋 Okami mid-event" "🧹 Clean base seed" \
      "💡 Light 24-fighter demo" "🔥 Dev hot reload" "🏗 Build Container Images" \
      "🛑 Stop containers" "♻️ Reset QA data" "← Back"
    choice=$?
    [ "$choice" -eq 255 ] && return
    case "$choice" in
      0) tui_log docker-full; spinner "Docker QA full tournament" bash scripts/qa-docker.sh full ;;
      1) tui_log docker-okami; spinner "Docker QA Okami mid-event" bash scripts/qa-docker.sh okami ;;
      2) tui_log docker-clean; spinner "Docker QA clean seed" bash scripts/qa-docker.sh clean ;;
      3) tui_log docker-light; spinner "Docker QA light demo" bash scripts/qa-docker.sh light ;;
      4) tui_log docker-dev; spinner "Docker QA dev hot reload" bash scripts/qa-docker.sh dev ;;
      5) tui_log docker-build; spinner "Building QA images" bash scripts/qa-docker.sh build ;;
      6) tui_log docker-down; spinner "Stopping QA containers" bash scripts/qa-docker.sh down ;;
      7) tui_log docker-reset; spinner "Resetting QA data" bash scripts/qa-docker.sh reset ;;
      8) tui_log docker-back; return ;;
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
      0) tui_log data-reset; spinner "Resetting demo data" go run ./cmd/j3d reset -db ./data/j3-demo.db ;;
      1) tui_log data-seed; spinner "Seeding demo data" go run ./cmd/j3d seed -db ./data/j3-demo.db ;;
      2) tui_log data-seed-full; spinner "Seeding full tournament" go run ./cmd/j3d reset -full -db ./data/j3-demo.db ;;
      3) tui_log data-run-live; if ! test_stub "run-live"; then J3_DB=./data/j3-live.db ./bin/j3d -addr 127.0.0.1:8787; fi ;;
      4) tui_log menu-back; return ;;
    esac
    pause
  done
}

build_and_deploy() {
  local environment="$1"
  if test_stub "build-and-deploy-$environment"; then return; fi
  spinner "Building $environment release (server + web + API + docs)" bash scripts/release.sh
  spinner "Deploying $environment" bash scripts/deploy.sh "$environment" dist
}

build_menu() {
  while true; do
    select_menu "Build & deploy" \
      "Build live executable + static UI" \
      "Build release (server + web + API + docs)" \
      "Run full CI/CD suite" "Deploy to development" "Deploy to staging" \
      "Deploy to production" "Build + deploy staging (one step)" \
      "Build + deploy production (one step)" "← Back"
    choice=$?
    [ "$choice" -eq 255 ] && return
    case "$choice" in
      0) tui_log build-exec; spinner "Building server + web into build/" bash scripts/local-build.sh local ;;
      1) tui_log build-release; spinner "Building server, web, API and docs release" bash scripts/release.sh ;;
      2) tui_log build-cicd; spinner "Full CI/CD suite" bash scripts/cicd.sh ;;
      3) tui_log deploy-dev; spinner "Deploying development" bash scripts/deploy.sh development dist ;;
      4) tui_log deploy-staging; spinner "Deploying staging" bash scripts/deploy.sh staging dist ;;
      5) tui_log deploy-prod; spinner "Deploying production" bash scripts/deploy.sh production dist ;;
      6) tui_log build-and-deploy-staging; build_and_deploy staging ;;
      7) tui_log build-and-deploy-production; build_and_deploy production ;;
      8) tui_log menu-back; return ;;
    esac
    pause
  done
}

build_hosted_apk() {
  if test_stub "build-hosted-apk"; then return; fi
  local url
  read -r -p "  HTTPS API URL (e.g. https://api.example.com): " url
  if [ -z "$url" ]; then echo "  ❌ No URL supplied"; return 1; fi
  J3_SERVER_URL="$url" bash scripts/mobile.sh build
}

mobile_menu() {
  while true; do
    select_menu "Mobile (Android / iOS)" \
      "Build demo APK" "Start Pixel 10 Pro emulator" "Install & launch APK" \
      "Build + install + launch" "Show phone server URL" "Run venue server for phone" \
      "Build iOS .app" "Build APK for hosted HTTPS server" "← Back"
    choice=$?
    [ "$choice" -eq 255 ] && return
    case "$choice" in
      0) tui_log mobile-build-apk; if ! test_stub "build-apk"; then bash scripts/mobile.sh build; fi ;;
      1) tui_log mobile-emulator; if ! test_stub "start-emulator"; then bash scripts/mobile.sh emulator; fi ;;
      2) tui_log mobile-install; if ! test_stub "install-apk"; then bash scripts/mobile.sh install; fi ;;
      3) tui_log mobile-build-install; if ! test_stub "build-install-launch"; then bash scripts/mobile.sh build && bash scripts/mobile.sh install; fi ;;
      4) tui_log mobile-url; echo "  📡 Phone URL: $(bash scripts/mobile.sh url)" ;;
      5) tui_log run-demo-phone; run_demo_phone ;;
      6) tui_log mobile-build-ios; if ! test_stub "build-ios"; then spinner "Building iOS app" bash scripts/mobile.sh build-ios; fi ;;
      7) tui_log mobile-build-hosted; build_hosted_apk ;;
      8) tui_log mobile-back; return ;;

    esac
    pause
  done
}

while true; do
  select_menu "J3 developer console" \
    "▶ Run full local demo" "🚀 Run fully seeded tournament" "🧪 QA & verification  ▸" \
    "💾 Demo data  ▸" "🏗 Build & Deploy  ▸" "📱 Mobile  ▸" \
    "🐳 Docker QA  ▸" "📚 Docs & guides  ▸" "❌ Exit"
  choice=$?
  case "$choice" in
    0) tui_log run-demo; run_demo ;;
    1) tui_log run-demo-full; run_demo_full ;;
    2) tui_log qa-menu; qa_menu ;;
    3) tui_log data-menu; data_menu ;;
    4) tui_log build-menu; build_menu ;;
    5) tui_log mobile-menu; mobile_menu ;;
    6) tui_log docker-menu; docker_menu ;;
    7) tui_log docs-menu; docs_menu ;;
    8) tui_log exit; echo "  👋 Bye!"; exit 0 ;;
  esac
done
