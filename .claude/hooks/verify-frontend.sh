#!/bin/bash
# Post-deploy hook: after web/ files are edited, build + deploy + run Playwright smoke test
# Only runs when frontend/UI files are modified

PROJECT_DIR="/Users/genglsh/workstation/cat/cat"
REMOTE_HOST="36.151.144.227"
REMOTE_USER="root"
REMOTE_PASS="Gls6622821-"
PROD_URL="http://${REMOTE_HOST}"

# Read hook input
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // .tool_response.filePath // ""' 2>/dev/null)

# Only trigger for web/ source files (not dist/)
if ! echo "$FILE_PATH" | grep -q "${PROJECT_DIR}/web/src/"; then
  exit 0
fi

# 1. Build frontend
cd "$PROJECT_DIR/web" && pnpm build:h5 > /dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "HOOK_ERROR: Frontend build failed"
  exit 1
fi

# 2. Deploy to remote
cd "$PROJECT_DIR/web/dist/build/h5"
tar czf /tmp/h5-hook.tar.gz .
sshpass -p "$REMOTE_PASS" scp -o StrictHostKeyChecking=no /tmp/h5-hook.tar.gz "${REMOTE_USER}@${REMOTE_HOST}:/tmp/h5-hook.tar.gz" 2>/dev/null
sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no "${REMOTE_USER}@${REMOTE_HOST}" \
  "cd /opt/cat/web && rm -rf assets static index.html && tar xzf /tmp/h5-hook.tar.gz && rm /tmp/h5-hook.tar.gz" 2>/dev/null
rm -f /tmp/h5-hook.tar.gz

# 3. Run Playwright smoke test
cd "$PROJECT_DIR/e2e"
RESULT=$(npx playwright test tests/smoke.spec.ts --reporter=line 2>&1)
EXIT_CODE=$?

if [ $EXIT_CODE -ne 0 ]; then
  echo "HOOK_ERROR: Playwright smoke test failed"
  echo "$RESULT" | tail -20
  exit 1
fi

echo "HOOK_OK: Frontend deployed and smoke test passed"
exit 0
