# 通知生成フロー図

この図は、`article.ingested` と `collector.fetch.failed` が RabbitMQ を経由して `notification-service` に取り込まれ、通知一覧 API と既読更新 API を通じて UI に見えるまでの流れを示します。
同期 API と非同期イベント処理の境界を分けて理解しやすくすることを目的にしています。

```mermaid
sequenceDiagram
  autonumber
  participant A as article-service 発行側
  participant C as collector-service 発行側
  participant MQ as RabbitMQ exchange<br/>tech-feed.events
  participant N as notification-service consumer
  participant DB as MySQL notifications
  participant G as api-gateway
  actor F as 利用者 / frontend

  alt article.ingested
    A->>MQ: article.ingested を publish
  else collector.fetch.failed
    C->>MQ: collector.fetch.failed を publish
  end

  MQ->>N: event を consumer に配送

  alt event_type = article.ingested
    N->>N: info 通知を組み立てる
    N->>DB: notifications 行を insert
  else event_type = collector.fetch.failed
    N->>N: error 通知を組み立てる
    N->>DB: notifications 行を insert
  end

  F->>G: 通知一覧を取得<br/>GET /api/v1/notifications
  G->>N: /api/v1/notifications を proxy
  N->>DB: notifications を取得
  DB-->>N: notification rows
  N-->>G: list response
  G-->>F: list response

  F->>G: 既読状態を更新<br/>PATCH /api/v1/notifications/:id/read-status
  G->>N: read-status update を proxy
  N->>DB: is_read と read_at を更新
  DB-->>N: updated row
  N-->>G: updated notification
  G-->>F: updated notification
```

## 補足

- `notification-service` は event の `event_type` を見て、`info` または `error` の通知レコードを生成します。
- UI の公開導線は `frontend -> api-gateway -> notification-service` で、RabbitMQ は UI から直接は見えません。
- 図では exchange と consumer を主題にしていますが、実装では `notification-service` queue を宣言し、`article.ingested` と `collector.fetch.failed` を bind しています。

## repo からは未確定

- dead-letter queue、max retry、複数 consumer、Slack / email / webhook などの配送先拡張は repo から確定できないため、この図には含めていません。

## 主な根拠ファイル

- `services/article-service/internal/events/publisher.go`
- `services/collector-service/internal/events/publisher.go`
- `services/notification-service/internal/events/contracts.go`
- `services/notification-service/internal/events/consumer.go`
- `services/notification-service/internal/service/notification_service.go`
- `services/api-gateway/internal/httpapi/routes.go`
- `docs/openapi/article.ingested.json`
- `docs/openapi/collector.fetch.failed.json`
