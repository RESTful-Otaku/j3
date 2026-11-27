#!/usr/bin/env bash
# Semver + J3 release-channel versioning.
# Usage: scripts/version.sh [alpha|beta|release] [get|check|name|code|strict|print]
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CHANNEL="${1:-release}"; ACTION="${2:-print}"

CORE="${J3_VERSION:-}"
if [ -z "$CORE" ]; then
  CORE="$(python3 -c 'import json;print(json.load(open("'"$ROOT"'/package.json"))["version"])')"
fi
CORE="${CORE#v}"; CORE="${CORE%%[-+]*}"
if ! printf '%s' "$CORE" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$'; then
  echo "❌ version '$CORE' is not strict semver x.y.z" >&2; exit 1
fi

case "$CHANNEL" in
  alpha)
    NAME="${CORE}a"          # user-facing J3 alpha (also strict ${CORE}-alpha)
    STRICT="${CORE}-alpha"   # SemVer pre-release
    CODE_OFFSET=0
    ;;
  beta)
    NAME="${CORE}b"          # user-facing J3 beta (also strict ${CORE}-beta)
    STRICT="${CORE}-beta"
    CODE_OFFSET=1
    ;;
  release|prod|production)
    NAME="$CORE"
    STRICT="$CORE"
    CODE_OFFSET=2
    ;;
  *) echo "unknown channel: $CHANNEL (alpha|beta|release)" >&2; exit 1;;
esac

major="${CORE%%.*}"; rest="${CORE#*.}"; minor="${rest%%.*}"; patch="${rest#*.}"; patch="${patch%%.*}"
# Monotonic Android/iOS build number: major*1000000 + minor*1000 + patch*10 + channel offset.
CODE=$(( major * 1000000 + minor * 1000 + patch * 10 + CODE_OFFSET ))

case "$ACTION" in
  check) echo "✅ semver ok: core=$CORE channel=$CHANNEL name=$NAME strict=$STRICT code=$CODE";;
  name) printf '%s\n' "$NAME";;
  strict) printf '%s\n' "$STRICT";;
  code) printf '%s\n' "$CODE";;
  core) printf '%s\n' "$CORE";;
  print|get)
    printf 'CORE=%s\nNAME=%s\nSTRICT=%s\nCODE=%s\nCHANNEL=%s\n' "$CORE" "$NAME" "$STRICT" "$CODE" "$CHANNEL"
    ;;
  *) echo "usage: scripts/version.sh [alpha|beta|release] [check|name|strict|code|core|print]" >&2; exit 1;;
esac
