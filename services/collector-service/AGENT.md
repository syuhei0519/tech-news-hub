# AGENT.md

## Read When

- RSS 取得、正規化、dedupe、ingest payload、source 同期を変えるとき
- 収集トリガーや publish 処理を変えるとき

## Open First

- `internal/collector/service.go`
- `internal/collector/service_test.go`
- `internal/collector/service_http_test.go`
- `internal/httpapi/routes.go`
- `internal/events/publisher.go`

## Usually Skip

- frontend の UI 実装
- notification-service の一覧 API

ただし ingest 契約を変える場合は article-service を同時に確認します。

## Minimum Verification

- `cd services/collector-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...`
