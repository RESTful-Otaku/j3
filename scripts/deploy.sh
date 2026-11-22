#!/usr/bin/env bash
# Deployment entrypoint for dev / staging / production.
# Override by setting DEPLOY_HOST, DEPLOY_USER and DEPLOY_PATH for rsync deploy.
set -euo pipefail
ENVIRONMENT="${1:-development}"
ARTIFACTS="${2:-artifacts}"

if [ ! -d "$ARTIFACTS" ]; then
  echo "❌ Artifact directory '$ARTIFACTS' not found"
  exit 1
fi

echo "🚀 Deploying to $ENVIRONMENT from $ARTIFACTS"
echo "   files: $(find "$ARTIFACTS" -type f | wc -l)"

if [ -n "${DEPLOY_HOST:-}" ]; then
  base_path="${DEPLOY_PATH:-/srv/j3}"
  if [[ "$base_path" == */"$ENVIRONMENT" ]]; then
    remote_path="$base_path"
  else
    remote_path="$base_path/$ENVIRONMENT"
  fi
  ssh_opts=(-o StrictHostKeyChecking=accept-new)
  if [ -n "${DEPLOY_SSH_KEY:-}" ]; then
    ssh_opts+=(-i "$DEPLOY_SSH_KEY")
  fi
  rsync -avz --delete -e "ssh ${ssh_opts[*]}" "$ARTIFACTS/" \
    "${DEPLOY_USER:-deploy}@${DEPLOY_HOST}:${remote_path}/"
  echo "✅ Deployed via rsync to $remote_path"
else
  echo "ℹ️  No DEPLOY_HOST configured — verified artifacts locally."
  echo "   Set DEPLOY_HOST/DEPLOY_USER/DEPLOY_PATH in CI secrets to enable remote deployment."
fi
