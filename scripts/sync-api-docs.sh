#!/usr/bin/env bash
# Keep the developer portal's copy of the API contract, guides and search index in sync.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="$ROOT/web/static/docs"
mkdir -p "$DEST"
cp "$ROOT/api/openapi.yaml" "$ROOT/web/static/openapi.yaml"
cp "$ROOT/docs/manifest.json" "$DEST/manifest.json"
cp "$ROOT"/docs/*.md "$DEST"/
python3 "$ROOT/scripts/build-docs-index.py" "$ROOT/api/openapi.yaml" "$DEST" "$ROOT/docs/manifest.json"
echo "✅ Synced API contract, guides and developer portal index"
