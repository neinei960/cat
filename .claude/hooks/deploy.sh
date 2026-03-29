#!/bin/bash
set -euo pipefail

# Auto-deploy hook: triggered after Edit/Write on server/ or web/ files.
# Uses a local lock to avoid overlapping deploys and performs safer remote swaps.

PROJECT_DIR="/Users/genglsh/workstation/cat/cat"
REMOTE_HOST="${REMOTE_HOST:-36.151.144.227}"
REMOTE_USER="${REMOTE_USER:-root}"
REMOTE_PASS="${REMOTE_PASS:-Gls6622821-}"
REMOTE_PORT="${REMOTE_PORT:-22}"
REMOTE_SSH_KEY="${REMOTE_SSH_KEY:-}"
LOCK_FILE="${TMPDIR:-/tmp}/cat-deploy.lock"
LOCK_DIR="${LOCK_FILE}.d"

cleanup_lock() {
  rm -rf "${LOCK_DIR}" 2>/dev/null || true
}

if command -v flock >/dev/null 2>&1; then
  exec 9>"${LOCK_FILE}"
  if ! flock -n 9; then
    echo "[deploy-hook] another deploy is already running, skipping"
    exit 0
  fi
else
  if ! mkdir "${LOCK_DIR}" 2>/dev/null; then
    echo "[deploy-hook] another deploy is already running, skipping"
    exit 0
  fi
  trap cleanup_lock EXIT
fi

ssh_opts=(-o StrictHostKeyChecking=no -p "${REMOTE_PORT}")
scp_opts=(-o StrictHostKeyChecking=no -P "${REMOTE_PORT}")

if [ -n "${REMOTE_SSH_KEY}" ] && [ -f "${REMOTE_SSH_KEY}" ]; then
  SSH_CMD=(ssh "${ssh_opts[@]}" -i "${REMOTE_SSH_KEY}" "${REMOTE_USER}@${REMOTE_HOST}")
  SCP_CMD=(scp "${scp_opts[@]}" -i "${REMOTE_SSH_KEY}")
else
  SSH_CMD=(sshpass -p "${REMOTE_PASS}" ssh "${ssh_opts[@]}" "${REMOTE_USER}@${REMOTE_HOST}")
  SCP_CMD=(sshpass -p "${REMOTE_PASS}" scp "${scp_opts[@]}")
fi

# Read the file path from stdin (hook input JSON)
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // .tool_response.filePath // ""' 2>/dev/null)

if [ -z "$FILE_PATH" ]; then
  exit 0
fi

IS_SERVER=false
IS_WEB=false

if echo "$FILE_PATH" | grep -q "${PROJECT_DIR}/server/"; then
  IS_SERVER=true
fi
if echo "$FILE_PATH" | grep -q "${PROJECT_DIR}/web/"; then
  IS_WEB=true
fi

if [ "$IS_SERVER" = false ] && [ "$IS_WEB" = false ]; then
  exit 0
fi

deploy_server() {
  local server_dir="${PROJECT_DIR}/server"
  local server_bin="${server_dir}/bin/server-linux"
  local remote_tmp="/opt/cat/server/server-new-$$-$(date +%s)"

  lsof -ti :8080 | xargs kill -9 2>/dev/null || true
  sleep 1
  (cd "${server_dir}" && go run cmd/server/main.go > /dev/null 2>&1 &) || true

  (cd "${server_dir}" && GOOS=linux GOARCH=amd64 go build -o "${server_bin}" cmd/server/main.go)

  "${SCP_CMD[@]}" "${server_bin}" "${REMOTE_USER}@${REMOTE_HOST}:${remote_tmp}"
  "${SSH_CMD[@]}" "
    set -euo pipefail
    cd /opt/cat/server
    chmod +x '${remote_tmp}'
    systemctl stop cat || true
    if [ -f server ]; then
      cp -f server server-bak
    fi
    mv -f '${remote_tmp}' server
    chmod +x server
    systemctl start cat
    ok=false
    for _ in 1 2 3 4 5; do
      if systemctl is-active --quiet cat && curl -fsS --max-time 3 http://127.0.0.1:8080/api/v1/health >/dev/null; then
        ok=true
        break
      fi
      sleep 1
    done
    if [ \"\$ok\" != true ]; then
      if [ -f server-bak ]; then
        cp -f server-bak server
        chmod +x server
      fi
      systemctl restart cat || true
      systemctl status cat --no-pager -l || true
      exit 1
    fi
  "
}

deploy_web() {
  local web_dir="${PROJECT_DIR}/web"
  local tarball
  local remote_tar="/tmp/cat-web-$$-$(date +%s).tar.gz"

  (cd "${web_dir}" && pnpm build:h5)

  tarball="$(mktemp /tmp/cat-web.XXXXXX).tar.gz"
  trap 'rm -f "${tarball}"' RETURN
  tar -C "${web_dir}/dist/build" -czf "${tarball}" h5

  "${SCP_CMD[@]}" "${tarball}" "${REMOTE_USER}@${REMOTE_HOST}:${remote_tar}"
  "${SSH_CMD[@]}" "
    set -euo pipefail
    rm -rf /opt/cat/web-next
    mkdir -p /opt/cat/web-next
    tar -xzf '${remote_tar}' -C /opt/cat/web-next --strip-components=1
    test -f /opt/cat/web-next/index.html
    rm -rf /opt/cat/web-prev
    if [ -d /opt/cat/web ]; then
      mv /opt/cat/web /opt/cat/web-prev
    fi
    mv /opt/cat/web-next /opt/cat/web
    rm -f '${remote_tar}'
  "
}

if [ "$IS_SERVER" = true ]; then
  deploy_server
fi

if [ "$IS_WEB" = true ]; then
  deploy_web
fi
