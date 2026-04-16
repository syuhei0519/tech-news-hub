# AGENT.md

## Responsibility

`notification-service` は通知の保存、配信契約、既読管理を担います。

- 通知一覧と既読更新 API を提供する
- RabbitMQ 経由の新着通知、取得失敗通知を取り込む
- 将来のダイジェストや外部通知連携の拡張点になる

## Read First

- `cmd/notification-service/main.go`: 起動エントリポイント
- `internal/app/app.go`: 初期化
- `internal/httpapi/router.go`: HTTP ルート
- `internal/service/notification_service.go`: ユースケース
- `internal/repository/notification_repository.go`: 永続化
- `internal/events/consumer.go`: イベント消費
- `internal/events/contracts.go`: イベント契約
- `internal/domain/models.go`: ドメインモデル

## Implementation Notes

- 通知 API の変更は `router.go` と `notification_service.go` を対で確認する
- イベント形式を変える場合は `contracts.go` を起点に producer 側も合わせる
- 既読更新や一覧取得の変更は repository のクエリ条件まで確認する
