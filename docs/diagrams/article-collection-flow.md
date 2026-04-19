# 記事収集フロー図

この図は、`collector-service` が source 設定を取得してから、fetch job を開始し、RSS を正規化して `article-service` に ingest し、最後に job を完了させるまでの現在実装を示します。
収集の主題は source 同期、逐次処理、job 管理、ingest 契約にあり、起動主体や将来の scheduler は図に含めていません。

```mermaid
sequenceDiagram
  autonumber
  actor Caller as 外部呼び出し元<br/>手動 / API-triggered collection
  participant Collector as collector-service
  participant Article as article-service
  participant RSS as 外部 RSS source
  participant DB as MySQL

  Caller->>Collector: POST /api/v1/collect/run
  Collector->>Article: GET /api/v1/sources
  Article->>DB: sources を取得
  DB-->>Article: source rows
  Article-->>Collector: items
  Note over Collector: disabled の source は除外する。<br/>enabled の source は順次処理する。

  loop enabled な source ごと
    Collector->>Article: POST /internal/fetch-jobs/start
    Article->>DB: fetch_jobs に running を記録
    Article-->>Collector: source_id, job_id

    Collector->>RSS: GET fetch_url

    alt RSS 取得と ingest が成功
      RSS-->>Collector: RSS feed
      Note over Collector: RSS item を正規化し、dedupe_key を生成する。
      Collector->>Article: POST /internal/ingest
      Article->>DB: articles を bulk upsert
      Article-->>Collector: inserted_count, duplicated_count
      Collector->>Article: POST /internal/fetch-jobs/:id/finish success
      Article->>DB: fetch_jobs と source status を更新
      Article-->>Collector: 204 No Content
    else RSS 取得または ingest が失敗
      Collector->>Article: POST /internal/fetch-jobs/:id/finish failed
      Article->>DB: fetch_jobs と source status を更新
      Article-->>Collector: 204 No Content
      Note over Collector: 現在実装では partial results を返し、<br/>最初のエラーで処理を止める。
    end
  end

  Collector-->>Caller: 202 Accepted または error response
```

## 補足

- `collector-service` は `GET /api/v1/sources` の結果から `is_enabled=true` の source だけを対象にします。
- `article-service` 側で `fetch_jobs` と `sources.last_fetch_status / last_error_message` を更新するため、source 詳細画面と job 履歴画面の整合が保たれます。
- `article-service` は ingest 成功で `inserted_count > 0` のとき `article.ingested` を publish しますが、この図では主題を収集フローに絞っています。

## repo からは未確定

- 定期実行主体、retry / backoff、並列収集ポリシーは repo から確定できないため、この図には含めていません。

## 主な根拠ファイル

- `services/collector-service/internal/httpapi/routes.go`
- `services/collector-service/internal/collector/service.go`
- `services/article-service/internal/httpapi/router.go`
- `services/article-service/internal/service/article_service.go`
- `docs/openapi/article-service-internal.yaml`
- `deployments/compose/mysql/init/001_schema.sql`
