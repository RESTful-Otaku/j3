#!/usr/bin/env bash
# Local full QA demo: one seeded venue server, one static web front end and a
# deterministic tournament snapshot. Uses a working Docker daemon when available,
# otherwise Podman; auto-selects free host ports when the defaults are busy.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

log() { printf '\n==> %s\n' "$*"; }
die() { echo "error: $*" >&2; exit 1; }

runtime_works() {
  command -v "$1" >/dev/null 2>&1 && "$1" info >/dev/null 2>&1
}

RUNTIME="${J3_CONTAINER_RUNTIME:-}"
if [[ -n "$RUNTIME" ]]; then
  runtime_works "$RUNTIME" || die "container runtime '$RUNTIME' is not running (check '${RUNTIME} info')"
else
  # A Docker CLI can be installed with no daemon; only pick runtimes that work.
  for candidate in docker podman; do
    if runtime_works "$candidate"; then RUNTIME="$candidate"; break; fi
  done
  if [[ -z "$RUNTIME" ]]; then
    echo "no running container runtime; trying scripts/doctor.sh…" >&2
    bash scripts/doctor.sh >&2 || true
    for candidate in docker podman; do
      if runtime_works "$candidate"; then RUNTIME="$candidate"; break; fi
    done
  fi
  [[ -n "$RUNTIME" ]] || die "no running container runtime found (tried docker and podman; see doctor output above)"
fi

port_available() {
  python3 - "$1" <<'PY'
import socket, sys
sock = socket.socket()
sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
try:
    sock.bind(('0.0.0.0', int(sys.argv[1])))
except OSError:
    raise SystemExit(1)
finally:
    sock.close()
PY
}

pick_port() {
  local start="$1" exclude="${2:-}" port
  for port in $(seq "$start" "$((start + 50))"); do
    [[ "$port" == "$exclude" ]] && continue
    if port_available "$port"; then printf '%s' "$port"; return 0; fi
  done
  die "no free port found from $start"
}

API_PORT="${J3_API_PORT:-8787}"
WEB_PORT="${J3_WEB_PORT:-5173}"

# Resolve only when starting a stack: stopped containers must not look like port conflicts.
resolve_ports() {
  if [[ -n "${J3_API_PORT:-}" ]] && ! port_available "$API_PORT"; then
    die "J3_API_PORT=$API_PORT is already in use"
  fi
  if [[ -n "${J3_WEB_PORT:-}" ]] && ! port_available "$WEB_PORT"; then
    die "J3_WEB_PORT=$WEB_PORT is already in use"
  fi
  if ! port_available "$API_PORT"; then
    local fallback
    fallback="$(pick_port 8787 "$WEB_PORT")"
    log "API port $API_PORT busy; using $fallback"
    API_PORT="$fallback"
  fi
  if ! port_available "$WEB_PORT" || [[ "$WEB_PORT" == "$API_PORT" ]]; then
    local fallback
    fallback="$(pick_port 5173 "$API_PORT")"
    log "Web port $WEB_PORT busy; using $fallback"
    WEB_PORT="$fallback"
  fi
}

SIM_SEED="${J3_SIM_SEED:-42}"
SIM_FIGHTERS="${J3_SIM_FIGHTERS:-50}"
NETWORK="j3-qa"
VOLUME="j3-qa-data"
API_NAME="j3-qa-api"
WEB_NAME="j3-qa-web"
API_IMAGE="j3-qa-api:local"
WEB_IMAGE="j3-qa-web:local"
WEB_DEV_IMAGE="j3-qa-web:dev"

container_exists() { "$RUNTIME" container inspect "$1" >/dev/null 2>&1; }
image_exists() { "$RUNTIME" image inspect "$1" >/dev/null 2>&1; }

published_port() {
  local name="$1" target="${2:-}" mapping
  if [[ -n "$target" ]]; then
    mapping="$("$RUNTIME" port "$name" "$target" 2>/dev/null | head -1 || true)"
  else
    mapping="$("$RUNTIME" port "$name" 2>/dev/null | head -1 || true)"
  fi
  [[ -n "$mapping" ]] || return 1
  printf '%s\n' "$mapping" | sed -E 's/.*:([0-9]+)[[:space:]]*$/\1/' | head -1
}

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
  local api_port web_port
  api_port="$(published_port "$API_NAME" 8787 || printf '%s' "$API_PORT")"
  web_port="$(published_port "$WEB_NAME" || printf '%s' "$WEB_PORT")"
  printf '\nJ3 QA demo is ready\n'
  printf '  Web:     http://localhost:%s\n' "$web_port"
  printf '  Replay:  http://localhost:%s/simulate\n' "$web_port"
  printf '  API:     http://localhost:%s/healthz\n' "$api_port"
  printf '  Runtime: %s\n' "$RUNTIME"
}

do_up() {
  local mode="$1"
  stop_stack
  resolve_ports
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

destroy_stack() {
  stop_stack
  log "Removing QA data volume + network + images"
  "$RUNTIME" volume rm -f "$VOLUME" >/dev/null 2>&1 || true
  "$RUNTIME" network rm "$NETWORK" >/dev/null 2>&1 || true
  "$RUNTIME" image rm -f "$API_IMAGE" "$WEB_IMAGE" "$WEB_DEV_IMAGE" >/dev/null 2>&1 || true
  log "QA stack destroyed (containers, volume, network, images)"
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
  destroy   Complete teardown: containers + volume + network + images
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
  destroy|teardown|clean-all) destroy_stack ;;
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
        state="$("$RUNTIME" inspect -f '{{.State.Status}}' "$name")"
        port="$(published_port "$name" || true)"
        if [[ -n "$port" ]]; then
          printf '%s: %s (host port %s)\n' "$name" "$state" "$port"
        else
          printf '%s: %s\n' "$name" "$state"
        fi
      else
        printf '%s: absent\n' "$name"
      fi
    done
    ;;
  url) print_urls ;;
  -h|--help|help) usage ;;
  *) usage; die "unknown command '$1'" ;;
esac
