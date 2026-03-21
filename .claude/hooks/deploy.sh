#!/bin/bash
# Auto-deploy hook: triggered after Edit/Write on server/ or web/ files
# Runs in background (async), does not block Claude

PROJECT_DIR="/Users/genglsh/workstation/cat/cat"
REMOTE_HOST="36.151.144.227"
REMOTE_USER="root"
REMOTE_PASS="Gls6622821-"

# Read the file path from stdin (hook input JSON)
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // .tool_response.filePath // ""' 2>/dev/null)

# Skip if no file path
if [ -z "$FILE_PATH" ]; then
  exit 0
fi

# Determine what changed
IS_SERVER=false
IS_WEB=false

if echo "$FILE_PATH" | grep -q "${PROJECT_DIR}/server/"; then
  IS_SERVER=true
fi
if echo "$FILE_PATH" | grep -q "${PROJECT_DIR}/web/"; then
  IS_WEB=true
fi

# Skip if neither server nor web source files
if [ "$IS_SERVER" = false ] && [ "$IS_WEB" = false ]; then
  exit 0
fi

# --- Deploy Server ---
if [ "$IS_SERVER" = true ]; then
  # 1. Restart local server
  lsof -ti :8080 | xargs kill -9 2>/dev/null || true
  sleep 1
  cd "$PROJECT_DIR/server" && go run cmd/server/main.go > /dev/null 2>&1 &

  # 2. Cross-compile for Linux
  cd "$PROJECT_DIR/server"
  GOOS=linux GOARCH=amd64 go build -o bin/server-linux cmd/server/main.go 2>/dev/null

  # 3. Upload and restart remote
  sshpass -p "$REMOTE_PASS" scp -o StrictHostKeyChecking=no "$PROJECT_DIR/server/bin/server-linux" "${REMOTE_USER}@${REMOTE_HOST}:/opt/cat/server/server-new" 2>/dev/null
  sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" 'cd /opt/cat/server && mv server server-bak 2>/dev/null; mv server-new server && chmod +x server && systemctl restart cat' 2>/dev/null
fi

# --- Deploy Web ---
if [ "$IS_WEB" = true ]; then
  # 1. Build frontend
  cd "$PROJECT_DIR/web" && pnpm build:h5 > /dev/null 2>&1

  # 2. Upload to remote
  sshpass -p "$REMOTE_PASS" scp -o StrictHostKeyChecking=no -r "$PROJECT_DIR/web/dist/build/h5/"* "${REMOTE_USER}@${REMOTE_HOST}:/opt/cat/web/" 2>/dev/null
fi

exit 0
