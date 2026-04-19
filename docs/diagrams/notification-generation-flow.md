# 通知生成フロー図

この図は、`article.ingested` と `collector.fetch.failed` が RabbitMQ を経由して `notification-service` に取り込まれ、通知一覧 API と既読更新 API を通じて UI に見えるまでの流れを示します。
同期 API と非同期イベント処理の境界を分けて理解しやすくすることを目的にしています。

```mermaid
sequenceDiagram
  autonumber
  participant A as article-service publisher
  participant C as collector-service publisher
  participant MQ as RabbitMQ exchange<br/>tech-feed.events
  participant N as notification-service consumer
  participant DB as MySQL notifications
  participant G as api-gateway
  actor F as frontend / Browser

  alt article.ingested
    A->>MQ: publish article.ingested
  else collector.fetch.failed
    C->>MQ: publish collector.fetch.failed
  end

  MQ->>N: deliver event to notification-service consumer

  alt event_type = article.ingested
    N->>N: build info notification
    N->>DB: insert notifications row
  else event_type = collector.fetch.failed
    N->>N: build error notification
    N->>DB: insert notifications row
  end

  F->>G: GET /api/v1/notifications
  G->>N: proxy /api/v1/notifications
  N->>DB: select notifications
  DB-->>N: notification rows
  N-->>G: list response
  G-->>F: list response

  F->>G: PATCH /api/v1/notifications/:id/read-status
  G->>N: proxy read-status update
  N->>DB: update is_read, read_at
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
