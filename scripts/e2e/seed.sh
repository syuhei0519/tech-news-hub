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

for _ in $(seq 1 30); do
  if docker compose exec -T mysql mysqladmin ping -h localhost -uroot "-p${MYSQL_ROOT_PASSWORD}" --silent >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

# E2E smoke は collector や RabbitMQ 到着待ちに寄せず、
# Playwright が前提にする最小 DB 状態を毎回 deterministic に再作成する。
docker compose exec -T mysql mysql -uroot "-p${MYSQL_ROOT_PASSWORD}" "${MYSQL_DATABASE}" < scripts/e2e/seed.sql
