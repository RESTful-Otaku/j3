#!/usr/bin/env bash
# Build release-ready artefacts locally (mirrors .github/workflows/release.yml).
set -euo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
# shellcheck source=build-paths.sh
source scripts/build-paths.sh
j3_prepare_build_dirs

echo "🧩 Building cross-platform server executables → $J3_BUILD_SERVER"
J3_VERSION="$(bash scripts/version.sh release core)" CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$J3_BUILD_SERVER/j3d-linux-amd64" ./cmd/j3d
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o "$J3_BUILD_SERVER/j3d-linux-arm64" ./cmd/j3d
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o "$J3_BUILD_SERVER/j3d-darwin-arm64" ./cmd/j3d
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o "$J3_BUILD_SERVER/j3d-windows-amd64.exe" ./cmd/j3d

# Release images use the production flavour; build:prod also mirrors build/web.
bash scripts/build-flavors.sh prod
test -f "$J3_BUILD_WEB/index.html"
test -f "$J3_BUILD_WEB/openapi.yaml"
test -f "$J3_BUILD_WEB/docs/manifest.json"
test -f "$J3_BUILD_WEB/docs/HOSTING.md"
test -f "$J3_BUILD_WEB/docs/hosting.html"
test -f "$J3_BUILD_WEB/LICENSE"

# Keep dist/ as a compatibility mirror for existing deploy commands and docs.
mkdir -p dist/bin dist/web
cp "$J3_BUILD_SERVER/j3d-linux-amd64" \
   "$J3_BUILD_SERVER/j3d-linux-arm64" \
   "$J3_BUILD_SERVER/j3d-darwin-arm64" \
   "$J3_BUILD_SERVER/j3d-windows-amd64.exe" dist/bin/
j3_stage_dir "$J3_BUILD_WEB" dist/web

echo "✅ Release candidates ready in $J3_BUILD_BASE (server + web + API + docs + licence)"
echo "   dist/ mirrors the server and web outputs for existing deploy commands"
ls -lh "$J3_BUILD_SERVER"
