#!/usr/bin/env bash
# Build release-ready artefacts locally (mirrors .github/workflows/release.yml).
set -euo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

mkdir -p dist/bin dist/web
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/bin/j3d-linux-amd64 ./cmd/j3d
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/bin/j3d-linux-arm64 ./cmd/j3d
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/bin/j3d-darwin-arm64 ./cmd/j3d
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/bin/j3d-windows-amd64.exe ./cmd/j3d
bun run sync:api-docs
bun run --cwd web build
cp -R web/build/. dist/web/
test -f dist/web/index.html
test -f dist/web/openapi.yaml
test -f dist/web/docs/manifest.json
test -f dist/web/docs/HOSTING.md
test -f dist/web/docs/hosting.html
test -f dist/web/LICENSE
echo "✅ Release artefacts ready in dist/ (server + web + API + docs + licence)"
ls -lh dist/bin