#!/usr/bin/env bash
# Preflight for Linux/WSL local + Docker QA + mobile real-device testing.
set -uo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'
ok() { printf "  ${GREEN}✔${NC} %s\n" "$1"; }
warn() { printf "  ${YELLOW}⚠${NC} %s\n" "$1"; }
fail() { printf "  ${RED}✖${NC} %s\n" "$1"; }
info() { printf "  ${CYAN}ℹ${NC} %s\n" "$1"; }
section() { printf "\n${BOLD}%s${NC}\n" "$1"; }

need() { command -v "$1" >/dev/null 2>&1; }
sudo_do() { if [ "$(id -u)" -eq 0 ]; then "$@"; elif need sudo; then sudo "$@"; else warn "need sudo for: $*"; return 1; fi; }

prompt_user() { printf "  ${YELLOW}→${NC} %s\n" "$1"; read -r -p "  Press Enter once done (or Ctrl-C to cancel)… " _; }

section "J3 doctor · local / Docker / mobile preflight (Linux/WSL)"
ok "repo: $ROOT"

# Basic tools
for tool in bash curl python3 git; do
  if need "$tool"; then ok "$tool"; else fail "$tool missing"; fi
done

# Go
if need go; then ok "go $(go version 2>/dev/null | awk '{print $3}')"; else
  warn "Go missing; trying install (linux apt/WSL)…"
  if need apt-get; then sudo_do apt-get update && sudo_do apt-get install -y golang-go || prompt_user "Install Go from https://go.dev/dl then re-run."; else prompt_user "Install Go from https://go.dev/dl then re-run."; fi
  need go && ok "go installed" || fail "go still missing"
fi

# Bun (no npm). Install user-local if missing.
if need bun; then ok "bun $(bun --version)"; else
  warn "Bun missing; installing user-local Bun…"
  if curl -fsSL https://bun.sh/install | bash; then export PATH="$HOME/.bun/bin:$PATH"; ok "bun installed"; else prompt_user "Install Bun from https://bun.sh then re-run."; fi
fi
if need bun; then ok "bun $(bun --version)"; fi

# Node/npm not required; used only by tooling that ships it.
need node && ok "node $(node --version)" || info "node not required"

# Docker or Podman
RUNTIME=""
for c in docker podman; do
  if need "$c"; then RUNTIME="$c"; break; fi
done
if [ -n "$RUNTIME" ]; then
  if "$RUNTIME" info >/dev/null 2>&1; then ok "$RUNTIME daemon ready"; else
    warn "$RUNTIME CLI present but daemon not running; trying start…"
    if [ "$RUNTIME" = docker ]; then
      sudo_do systemctl start docker 2>/dev/null || sudo_do service docker start 2>/dev/null || true
      prompt_user "If Docker Desktop/WSL integration isn't running, start it, then press Enter."
    else
      prompt_user "Start Podman (podman machine start / systemctl start podman), then press Enter."
    fi
    "$RUNTIME" info >/dev/null 2>&1 && ok "$RUNTIME daemon ready" || fail "$RUNTIME daemon still not ready"
  fi
else
  warn "No container runtime; trying install Docker…"
  if need apt-get; then sudo_do apt-get update && sudo_do apt-get install -y docker.io || prompt_user "Install Docker/Podman then re-run."; sudo_do systemctl start docker 2>/dev/null || true; else prompt_user "Install Docker/Podman then re-run."; fi
fi

# Web deps
if [ -f web/bun.lock ]; then info "bun install --cwd web --frozen-lockfile recommended"; fi

# Mobile real device
need java && ok "java" || info "Java 21 needed for Android"
need adb && ok "adb" || info "Android SDK platform-tools needed for real device QA"
need keytool && ok "keytool" || info "JDK keytool needed for Android signing"
need xcodebuild && ok "xcodebuild" || info "Xcode on macOS needed for real iOS; Linux/WSL builds Android only"

# Env templates
if [ ! -f .env.local ]; then info "copy .env.local.example -> .env.local"; fi

section "Done. Re-run 'bash scripts/doctor.sh' after any fix."
