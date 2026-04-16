# Agent Playbooks

## Purpose

このドキュメントは、よくある作業ごとに最小読込セットを固定するためのものです。
作業開始時は、必要なセクションだけ読みます。

## Common Rule

どの作業でも最初に確認するもの:

- `AGENT.md`
- `docs/current-status.md`
- 対象領域の `AGENT.md`

必要になってから開くもの:

- `docs/openapi/api-gateway.summary.md`
- `RULES.md`
- 領域別 guideline
- `docs/runbook.md`
- `docs/test-policy.md`

## Docs-Only

### Read

- `AGENT.md`
- 更新対象の docs

### Usually Skip

- 実装コード
- 重い検証手順

### Verification

- リンク先、コマンド、ファイル名の整合だけ確認する
- PR を作る場合は `.github/pull_request_template.md` を開き、GitHub API / MCP 作成時も本文を手動で組み立てる

## Frontend UI Change

### Read

- `frontend/AGENT.md`
- 対象画面
- `frontend/src/lib/api.ts`

### Read If Needed

- `frontend/src/store/searchStore.ts`
- `services/api-gateway/AGENT.md`
- `docs/openapi/api-gateway.summary.md`
- `docs/openapi/api-gateway.yaml`

### Verification

- `cd frontend && npm run build`

## Public API Change

### Read

- `services/api-gateway/AGENT.md`
- 対象 upstream service の `AGENT.md`
- `docs/openapi/api-gateway.summary.md`
- `docs/api-guidelines.md`

### Read If Needed

- `frontend/AGENT.md`
- `RULES.md`
- `docs/openapi/api-gateway.yaml`

### Verification

- 対象 Go service の `go test ./...`
- 影響が複数領域に跨るなら `make verify`

## Article-Service Change

### Read

- `services/article-service/AGENT.md`
- 対象 `service` / `repository` / `router`

### Read If Needed

- `docs/db-guidelines.md`
- `docs/test-policy.md`
- `services/collector-service/AGENT.md`
- `services/notification-service/AGENT.md`

### Verification

- `cd services/article-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...`

## Collector Change

### Read

- `services/collector-service/AGENT.md`
- 対象 `service.go` とテスト

### Read If Needed

- `services/article-service/AGENT.md`
- `docs/test-policy.md`
- `docs/runbook.md`

### Verification

- `cd services/collector-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...`

## Notification Change

### Read

- `services/notification-service/AGENT.md`
- 対象 `service` / `repository` / `events`

### Read If Needed

- producer 側 service の `AGENT.md`
- `docs/test-policy.md`
- `docs/runbook.md`

### Verification

- `cd services/notification-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...`

## DB / Repository Change

### Read

- 対象 service の `AGENT.md`
- `docs/db-guidelines.md`
- `deployments/compose/mysql/init/001_schema.sql`

### Read If Needed

- `RULES.md`
- `docs/release-migration-policy.md`

### Verification

- 対象 service の `go test ./...`
- 変更が広い場合は `make verify`

## Compose / CI / Infra Change

### Read

- `AGENT.md`
- `docs/runbook.md`
- `docs/test-policy.md`

### Read If Needed

- `docs/k8s-guidelines.md`
- `docs/branch-strategy.md`
- `RULES.md`

### Verification

- 対象コマンド
- 迷う場合は `make verify`

## Rule / Policy Change

### Read

- `RULES.md`
- `docs/rules-operations.md`
- `AGENT.md`

### Read If Needed

- `docs/requirements.md`
- `docs/dod.md`
- `docs/branch-strategy.md`

### Verification

- 変更された方針が矛盾なく参照できるかを確認する

## Bugfix Triage

### Read

- 発生箇所の `AGENT.md`
- 失敗しているコード
- 直近で関連するテスト

### Read If Needed

- `docs/runbook.md`
- `docs/known-issues.md`
- `docs/test-policy.md`

### Verification

- 再現に近い最小テスト
- 必要なら回帰テストを追加
