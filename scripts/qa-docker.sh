#!/usr/bin/env bash
# Local full QA demo: one seeded venue server, one static web front end and a
# deterministic tournament snapshot. Uses Docker when available, Podman when not.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

RUNTIME="${J3_CONTAINER_RUNTIME:-}"
if [[ -z "$RUNTIME" ]]; then
  if command -v docker >/dev/null 2>&1; then RUNTIME=docker
  elif command -v podman >/dev/null 2>&1; then RUNTIME=podman
  else
    echo "No container runtime found. Install Docker or Podman." >&2
    exit 1
  fi
fi

API_PORT="${J3_API_PORT:-8787}"
WEB_PORT="${J3_WEB_PORT:-5173}"
SIM_SEED="${J3_SIM_SEED:-42}"
SIM_FIGHTERS="${J3_SIM_FIGHTERS:-50}"
NETWORK="j3-qa"
VOLUME="j3-qa-data"
API_NAME="j3-qa-api"
WEB_NAME="j3-qa-web"
API_IMAGE="j3-qa-api:local"
WEB_IMAGE="j3-qa-web:local"
WEB_DEV_IMAGE="j3-qa-web:dev"

log() { printf '\n==> %s\n' "$*"; }
die() { echo "error: $*" >&2; exit 1; }

container_exists() { "$RUNTIME" container inspect "$1" >/dev/null 2>&1; }
image_exists() { "$RUNTIME" image inspect "$1" >/dev/null 2>&1; }

ensure_network() {
  if ! "$RUNTIME" network inspect "$NETWORK" >/dev/null 2>&1; then
    log "Creating network $NETWORK"
    "$RUNTIME" network create "$NETWORK" >/dev/null
  fi
}

ensure_volume() {
  if ! "$RUNTIME" volume inspect "$VOLUME" >/dev/null 2>&1; then
    log "Creating volume $VOLUME"
    "$RUNTIME" volume create "$VOLUME" >/dev/null
  fi
}

ensure_api_image() {
  local force="${1:-}"
  if [[ "$force" == force ]] || ! image_exists "$API_IMAGE"; then
    log "Building venue server image"
    "$RUNTIME" build -t "$API_IMAGE" -f Dockerfile .
  fi
}

ensure_web_image() {
  local target="${1:-prod}" force="${2:-}" image="$WEB_IMAGE"
  [[ "$target" == dev ]] && image="$WEB_DEV_IMAGE"
  if [[ "$force" == force ]] || ! image_exists "$image"; then
    log "Building web image ($target)"
    "$RUNTIME" build -t "$image" -f web/Dockerfile --target "$target" .
  fi
}

stop_stack() {
  for name in "$WEB_NAME" "$API_NAME"; do
    if container_exists "$name"; then
      log "Removing container $name"
      "$RUNTIME" rm -f "$name" >/dev/null
    fi
  done
}

reset_volume() {
  if "$RUNTIME" volume inspect "$VOLUME" >/dev/null 2>&1; then
    log "Resetting volume $VOLUME"
    "$RUNTIME" volume rm -f "$VOLUME" >/dev/null
  fi
  ensure_volume
}

run_one_shot() {
  "$RUNTIME" run --rm --network "$NETWORK" -v "$VOLUME:/data" "$API_IMAGE" "$@"
}

seed_for_mode() {
  local mode="$1" fighters="$SIM_FIGHTERS"
  local args=()
  case "$mode" in
    full|up|dev)
      args=(reset -okami -full -db /data/j3.db)
      ;;
    okami)
      args=(reset -okami -full=false -db /data/j3.db)
      ;;
    clean|light)
      args=(reset -okami=false -full=false -db /data/j3.db)
      fighters=24
      ;;
    *) die "unknown mode '$mode'" ;;
  esac
  log "Seeding $mode tournament"
  run_one_shot "${args[@]}"
  log "Generating deterministic snapshot (seed $SIM_SEED, $fighters fighters)"
  run_one_shot simulate -seed "$SIM_SEED" -fighters "$fighters" -out /data/simulation.json
}

start_api() {
  log "Starting venue API on :$API_PORT"
  "$RUNTIME" run -d --name "$API_NAME" --network "$NETWORK" \
    -p "$API_PORT:8787" -v "$VOLUME:/data" \
    -e J3_ADDR=0.0.0.0:8787 -e J3_DB=/data/j3.db \
    -e J3_DEMO=1 -e J3_SEED_DEMO=0 -e J3_VERSION=demo \
    "$API_IMAGE" >/dev/null
}

start_web() {
  local mode="$1"
  if [[ "$mode" == dev ]]; then
    log "Starting Vite dev web on :$WEB_PORT"
    "$RUNTIME" run -d --name "$WEB_NAME" --network "$NETWORK" -p "$WEB_PORT:5173" \
      -e VITE_API_TARGET="http://$API_NAME:8787" \
      -v "$ROOT/web:/app:z" -v /app/node_modules \
      "$WEB_DEV_IMAGE" bun run dev >/dev/null
  else
    log "Starting static web + API proxy on :$WEB_PORT"
    "$RUNTIME" run -d --name "$WEB_NAME" --network "$NETWORK" -p "$WEB_PORT:8080" \
      "$WEB_IMAGE" >/dev/null
  fi
}

wait_ready() {
  log "Waiting for the venue API"
  local i
  for i in $(seq 1 60); do
    if curl -fsS "http://127.0.0.1:$API_PORT/healthz" >/dev/null 2>&1; then return 0; fi
    sleep 1
  done
  "$RUNTIME" logs "$API_NAME" >&2 || true
  die "API did not become ready on port $API_PORT"
}

print_urls() {
  printf '\nJ3 QA demo is ready\n'
  printf '  Web:     http://localhost:%s\n' "$WEB_PORT"
  printf '  Replay:  http://localhost:%s/simulate\n' "$WEB_PORT"
  printf '  API:     http://localhost:%s/healthz\n' "$API_PORT"
  printf '  Runtime: %s\n' "$RUNTIME"
}

do_up() {
  local mode="$1"
  stop_stack
  ensure_api_image force
  if [[ "$mode" == dev ]]; then ensure_web_image dev force; else ensure_web_image prod force; fi
  ensure_network
  reset_volume
  seed_for_mode "$mode"
  start_api
  start_web "$mode"
  wait_ready
  print_urls
}

usage() {
  cat <<EOF
J3 local QA demo (Docker with Podman fallback)

Usage: bash scripts/qa-docker.sh <command>

Modes (recreate the stack with a fresh seeded database):
  up|full   Complete tournament + mid-event Okami seed (default)
  okami     Mid-event Okami 2027 seed only
  clean     Minimal base seed
  light     Minimal seed with a small 24-fighter snapshot
  dev       Full seed with Vite hot reload mounted from ./web

Lifecycle:
  build     Build the API and static web images
  down      Stop and remove the containers
  reset     Stop, then discard the database/simulation volume
  simulate  Regenerate /data/simulation.json without restarting
  logs      Show recent API logs
  follow    Stream API logs
  status    Show container state
  url       Print the demo URLs

Environment: J3_API_PORT, J3_WEB_PORT, J3_SIM_SEED, J3_SIM_FIGHTERS,
             J3_CONTAINER_RUNTIME (docker|podman)
EOF
}

case "${1:-up}" in
  up|full|okami|clean|light|dev) do_up "$1" ;;
  build) ensure_api_image force; ensure_web_image prod force; ensure_web_image dev force ;;
  down) stop_stack; log "QA stack stopped" ;;
  reset) stop_stack; reset_volume; log "QA data reset" ;;
  simulate)
    ensure_api_image; ensure_network; ensure_volume
    run_one_shot simulate -seed "$SIM_SEED" -fighters "$SIM_FIGHTERS" -out /data/simulation.json
    log "Snapshot regenerated"
    ;;
  logs)
    if container_exists "$API_NAME"; then "$RUNTIME" logs --tail 100 "$API_NAME"; else echo "API is not running"; fi
    ;;
  follow)
    container_exists "$API_NAME" || die "API is not running"
    "$RUNTIME" logs -f "$API_NAME"
    ;;
  status)
    for name in "$API_NAME" "$WEB_NAME"; do
      if container_exists "$name"; then
        printf '%s: %s\n' "$name" "$("$RUNTIME" inspect -f '{{.State.Status}}' "$name")"
      else
        printf '%s: absent\n' "$name"
      fi
    done
    ;;
  url) print_urls ;;
  -h|--help|help) usage ;;
  *) usage; die "unknown command '$1'" ;;
esac
