#!/usr/bin/env bash
# Build explicit local dev, QA/staging and production web bundles. Each flavour
# is staged under build/web-<flavour>; production is also mirrored into
# build/web and dist/web so existing deploy commands keep working.
set -euo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
# shellcheck source=build-paths.sh
source scripts/build-paths.sh
j3_prepare_build_dirs

build_flavor() {
  local flavor="$1"
  local dest="$J3_BUILD_BASE/web-$flavor"
  local api_url="" offline=0 autologin=0 full_phone=0
  case "$flavor" in
    dev)
      api_url="${J3_DEV_API_URL:-}"
      ;;
    qa)
      # QA is a self-contained bundle with the bundled seeded database and API
      # snapshot, so it needs no server and works from any static host.
      offline=1
      autologin=1
      full_phone=1
      ;;
    prod)
      api_url="${J3_PROD_API_URL:-${J3_SERVER_URL:-}}"
      ;;
    *) echo "Unknown web flavour: $flavor" >&2; exit 1 ;;
  esac

  echo "🌐 Building $flavor web bundle → $dest"
  VITE_API_BASE_URL="$api_url" \
  VITE_BUILD_FLAVOR="$flavor" \
  VITE_OFFLINE_DEMO="$offline" \
  VITE_QA_AUTOLOGIN="$autologin" \
  VITE_QA_FULL_PHONE="$full_phone" \
    bun run --cwd web build >/dev/null
  j3_stage_dir web/build "$dest"
  echo "✅ $flavor web build ready: $dest"
}

go build ./...
bun run sync:api-docs >/dev/null

case "${1:-all}" in
  dev|qa|prod) build_flavor "$1" ;;
  all)
    build_flavor dev
    build_flavor qa
    build_flavor prod
    ;;
  *) echo "Usage: scripts/build-flavors.sh {dev|qa|prod|all}" >&2; exit 1 ;;
esac

if [[ "${1:-all}" == "all" || "${1:-all}" == "prod" ]]; then
  j3_stage_dir "$J3_BUILD_BASE/web-prod" "$J3_BUILD_WEB"
  j3_stage_dir "$J3_BUILD_BASE/web-prod" dist/web
  echo "✅ Production web mirrored to $J3_BUILD_WEB and dist/web"
fi
