# システム全体構成図

この図は、現在の runtime 構成、公開 API 境界、内部連携、データ / メッセージ基盤を 1 枚で把握するためのものです。
新規参入者や AI エージェントが、どの service がどこに依存し、どの経路が同期 API でどの経路が非同期イベントかを素早く掴めるようにしています。

```mermaid
flowchart LR
  subgraph Client[クライアント]
    user[利用者 / Browser]
    frontend[frontend<br/>UI<br/>React + TypeScript + Vite]
  end

  subgraph Public_API[公開 API]
    gateway[api-gateway<br/>公開 API /api/v1/*]
  end

  subgraph Processing[処理系 service]
    article[article-service]
    collector[collector-service]
    notification[notification-service]
  end

  subgraph Data_Messaging[データ / メッセージ基盤]
    mysql[(MySQL<br/>articles / sources / fetch_jobs / notifications)]
    rabbit[(RabbitMQ exchange<br/>tech-feed.events)]
  end

  subgraph External[外部]
    rss[外部 RSS sources]
  end

  user --> frontend
  frontend -->|HTTP /api/v1/*| gateway
  gateway -->|記事 / source / fetch-jobs / CSV| article
  gateway -->|通知一覧 / 既読更新| notification

  collector -->|source 一覧取得<br/>GET /api/v1/sources| article
  collector -->|fetch job 開始<br/>POST /internal/fetch-jobs/start| article
  collector -->|ingest<br/>POST /internal/ingest| article
  collector -->|fetch job 完了<br/>POST /internal/fetch-jobs/:id/finish| article
  collector -->|RSS 取得<br/>GET fetch_url| rss

  article -->|読み書き| mysql
  notification -->|読み書き| mysql

  article -->|article.ingested を publish| rabbit
  collector -->|collector.fetch.failed を publish| rabbit
  rabbit -->|event を consume| notification
```

## 補足

- `collector-service` は `frontend` や `api-gateway` の公開導線には入っておらず、別入口の収集実行 service として存在します。
- `collector-service` は DB を直接触らず、`article-service` の public / internal API を使います。
- `notification-service` は RabbitMQ の consumer であると同時に、通知一覧 / 既読更新 API の upstream でもあります。

## repo からは未確定

- `collector-service` の通常運用での起動主体は repo から確定できないため、この図では表現していません。

## 主な根拠ファイル

- `README.md`
- `docker-compose.yml`
- `services/api-gateway/internal/httpapi/routes.go`
- `services/article-service/internal/app/app.go`
- `services/collector-service/internal/app/app.go`
- `services/collector-service/internal/collector/service.go`
- `services/notification-service/internal/app/app.go`
- `services/notification-service/internal/events/consumer.go`
- `deployments/compose/mysql/init/001_schema.sql`
