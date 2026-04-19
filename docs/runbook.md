# Runbook

## Purpose

このドキュメントは、開発や障害切り分けで頻繁に使う基本運用手順をまとめたもの。

## Local Startup

### Initial Setup

```bash
cp .env.example .env
docker compose config -q
```

補足:

- RabbitMQ 管理 UI は `http://localhost:15672`
- 認証情報は `.env` の `RABBITMQ_USER`, `RABBITMQ_PASSWORD`

### Start

```bash
make up
```

### Start With Hot Reload

```bash
make dev-up
```

補足:

- frontend は Vite HMR で即時反映する
- Go services は `air` で build / restart される
- 初回起動では dev image build と依存取得に時間がかかる
- dev 用 compose は `docker-compose.yml` に `docker-compose.dev.yml` を重ねて使う
- frontend 依存が変わった場合も `package-lock.json` を見て `npm ci` を再実行する

### Stop

```bash
make down
```

### Stop Hot Reload Stack

```bash
make dev-down
```

補足:

- `make down` は named volume を削除しない
- DB を含めて初期化したい場合だけ `make reset` を使う

### Logs

```bash
make logs
```

### Dev Logs

```bash
make dev-logs
```

## Basic Verification

### Go Services

```bash
cd services/article-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
cd services/api-gateway && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
cd services/collector-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
cd services/notification-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
```

### article-service MySQL Integration Test

```bash
docker compose up -d mysql
ARTICLE_SERVICE_TEST_MYSQL_DSN='app_user:app_password@tcp(127.0.0.1:3306)/tech_feed_hub?parseTime=true&multiStatements=true' make test-article-integration
```

補足:

- integration test は `ARTICLE_SERVICE_RUN_INTEGRATION=1` が付いたときだけ実行される
- `ARTICLE_SERVICE_TEST_MYSQL_DSN` を省略した場合は `127.0.0.1:3306` のローカル MySQL を既定で参照する
- schema は `deployments/compose/mysql/init/001_schema.sql` をそのまま適用する

### Frontend

```bash
cd frontend
npm install
npm run test:run
npm run build
```

または:

```bash
make test-frontend
```

### Compose Config

```bash
docker compose config -q
```

### Dev Compose Config

```bash
make dev-config
```

### Full Verify

```bash
make verify
```

補足:

- `make verify` は backend test + frontend test + build + compose config check をまとめて実行する
- MySQL が必要な article-service integration test は `make verify` に含めない

## CI Lanes

- `verify`: `make verify` を実行する高速レーン
- `integration`: article-service の MySQL integration test を実行するレーン
- `smoke`: Playwright E2E を relevant changes 時に実行するレーン

補足:

- docs-only 変更では `verify` は required check を維持しつつ重い検証を省略する
- `integration` と `smoke` は docs-only 変更では実行しない

## Health Endpoints

- `api-gateway`: `/health`
- `article-service`: `/health`
- `collector-service`: `/health`
- `notification-service`: `/health`

## Common Troubleshooting

### Frontend Build Failure

確認すること:

- `npm install` が最新か
- `tsconfig` と `vite` 設定変更が入っていないか
- API 型変更に frontend が追従しているか

### Article List Not Showing Data

確認すること:

- `article-service` が起動しているか
- `api-gateway` の upstream 設定が正しいか
- MySQL が起動しているか
- articles テーブルにデータが入っているか

### Collector Ingest Failure

確認すること:

- 外部ソース URL にアクセスできるか
- source の設定が正しいか
- `article-service` の `/internal/ingest` が生きているか
- dedupe / DB 制約で失敗していないか
- fetch_jobs と source status が更新されているか
- UI では `/sources/:id` のジョブ履歴画面で失敗メッセージを確認できるか
- UI では `/notifications` に failure 通知が出ているか

### Notification Delay / Missing

確認すること:

- RabbitMQ が起動しているか
- `notification-service` が起動しているか
- RabbitMQ 管理 UI で queue / message が滞留していないか
- `notifications` テーブルにレコードが入っているか
- `article-service`, `collector-service` の publish error ログが出ていないか

### CSV Export Mismatch

確認すること:

- 一覧 API と同じフィルタ条件を使っているか
- 並び順と日付フォーマットが固定されているか
- 件数上限に引っかかっていないか

## Incident Triage Order

障害時は次の順で見る。

1. `api-gateway` health
2. `article-service` health
3. MySQL 状態
4. collector-service 状態
5. notification-service 状態

理由:

- 記事閲覧継続性を最優先するため

## Before Merging

- 関連 docs を更新したか
- OpenAPI 更新が必要か確認したか
- schema 変更があるか確認したか
- compose 構成を壊していないか確認したか

## Future Runbook Topics

後で追加する。

- kind デプロイ手順
- Helm 更新手順
- Argo CD 同期確認手順
- Proxmox クラスタ運用手順
- バックアップ / リストア手順
