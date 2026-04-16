# AGENT.md

## Read When

- 記事、ソース、fetch job の振る舞いを変えるとき
- ingest、検索、CSV 前提データ、永続化条件を変えるとき

## Open First

- `internal/httpapi/router.go`
- `internal/service/article_service.go`
- `internal/repository/`
- `internal/domain/models.go`
- `internal/events/publisher.go`

## Usually Skip

- frontend の見た目
- notification-service の内部実装

ただし ingest 契約変更時は collector、通知イベント変更時は notification も確認します。

## Minimum Verification

- `cd services/article-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...`
