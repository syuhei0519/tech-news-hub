#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${ROOT_DIR}"

if [[ ! -f .env ]]; then
  echo ".env が見つかりません。.env.example をコピーしてから実行してください。" >&2
  exit 1
fi

set -a
source .env
set +a

wait_for_url() {
  local url="$1"
  local label="$2"

  for _ in $(seq 1 60); do
    if curl -fsS "${url}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done

  echo "${label} の起動待ちがタイムアウトしました: ${url}" >&2
  exit 1
}

# compose の depends_on だけでは host からの利用可能時点までは保証しないため、E2E 前に明示的に待機する。
wait_for_url "http://127.0.0.1:${ARTICLE_SERVICE_PORT}/health" "article-service"
wait_for_url "http://127.0.0.1:${NOTIFICATION_SERVICE_PORT}/health" "notification-service"
wait_for_url "http://127.0.0.1:${API_GATEWAY_PORT}/health" "api-gateway"
wait_for_url "http://127.0.0.1:${FRONTEND_PORT}" "frontend"
