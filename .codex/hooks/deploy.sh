#!/bin/bash
set -euo pipefail

# Codex wrapper around the existing Claude deployment hook.
# Input format matches the Claude hook JSON so both can share the same script.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CLAUDE_DEPLOY_SCRIPT="${PROJECT_DIR}/.claude/hooks/deploy.sh"

if [ ! -x "${CLAUDE_DEPLOY_SCRIPT}" ]; then
  chmod +x "${CLAUDE_DEPLOY_SCRIPT}"
fi

exec "${CLAUDE_DEPLOY_SCRIPT}"
