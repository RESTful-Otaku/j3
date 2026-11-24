#!/usr/bin/env bash
# Build local web/server targets into build/{web,server}. Mobile helpers stage
# their own outputs into build/android and build/ios.
set -euo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
# shellcheck source=build-paths.sh
source scripts/build-paths.sh
j3_prepare_build_dirs

build_web() {
  echo "🌐 Building web bundle → $J3_BUILD_WEB"
  go build ./...
  bun run sync:api-docs
  bun run --cwd web build
  j3_stage_dir web/build "$J3_BUILD_WEB"
  test -f "$J3_BUILD_WEB/index.html"
  echo "✅ Web build ready: $J3_BUILD_WEB"
}

build_server() {
  echo "🧩 Building local server executable → $J3_BUILD_SERVER/j3d"
  CGO_ENABLED=0 go build -o "$J3_BUILD_SERVER/j3d" ./cmd/j3d
  echo "✅ Server executable ready: $J3_BUILD_SERVER/j3d"
}

build_android() {
  bash scripts/mobile.sh build
}

build_ios() {
  bash scripts/mobile.sh build-ios
}

case "${1:-all}" in
  web) build_web ;;
  server) build_server ;;
  local) build_server; build_web ;;
  android) build_android ;;
  ios) build_ios ;;
  all)
    build_server
    build_web
    build_android
    build_ios
    ;;
  *)
    echo "Usage: scripts/local-build.sh {web|server|local|android|ios|all}" >&2
    exit 1
    ;;
esac
