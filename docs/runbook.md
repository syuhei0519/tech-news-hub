# Runbook

## Purpose

このドキュメントは、開発や障害切り分けで頻繁に使う基本運用手順をまとめたもの。

## Local Startup

### Initial Setup

```bash
cp .env.example .env
docker compose config -q
```

### Start

```bash
make up
```

### Stop

```bash
make down
```

### Logs

```bash
make logs
```

## Basic Verification

### Go Services

```bash
cd services/article-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
cd services/api-gateway && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
cd services/collector-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
cd services/notification-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
```

### Frontend

```bash
cd frontend
npm install
npm run build
```

### Compose Config

```bash
docker compose config -q
```

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
