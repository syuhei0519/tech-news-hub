# AGENT.md

## Read When

- 通知一覧、既読更新、イベント消費、通知保存を変えるとき
- 新着通知や取得失敗通知の契約を変えるとき

## Open First

- `internal/httpapi/router.go`
- `internal/service/notification_service.go`
- `internal/repository/notification_repository.go`
- `internal/events/consumer.go`
- `internal/events/contracts.go`

## Usually Skip

- frontend の画面実装
- collector の収集詳細

ただしイベント契約変更時は producer 側の service も確認します。

## Minimum Verification

- `cd services/notification-service && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...`
