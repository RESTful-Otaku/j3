#!/usr/bin/env bash
# Keep the developer portal's copy of the OpenAPI contract in sync.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cp "$ROOT/api/openapi.yaml" "$ROOT/web/static/openapi.yaml"
mkdir -p "$ROOT/web/static/docs"
cp "$ROOT/docs/API.md" "$ROOT/web/static/docs/API.md"
echo "✅ Synced API contract and developer guide"
