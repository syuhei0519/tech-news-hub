# 記事収集フロー図

この図は、`collector-service` が source 設定を取得してから、fetch job を開始し、RSS を正規化して `article-service` に ingest し、最後に job を完了させるまでの現在実装を示します。
収集の主題は source 同期、逐次処理、job 管理、ingest 契約にあり、起動主体や将来の scheduler は図に含めていません。

```mermaid
sequenceDiagram
  autonumber
  actor Caller as External caller<br/>manual / API-triggered collection
  participant Collector as collector-service
  participant Article as article-service
  participant RSS as External RSS source
  participant DB as MySQL

  Caller->>Collector: POST /api/v1/collect/run
  Collector->>Article: GET /api/v1/sources
  Article->>DB: read sources
  DB-->>Article: source rows
  Article-->>Collector: items
  Note over Collector: Disabled sources are skipped.<br/>Enabled sources are processed sequentially.

  loop each enabled source
    Collector->>Article: POST /internal/fetch-jobs/start
    Article->>DB: insert fetch_jobs(status=running)
    Article-->>Collector: source_id, job_id

    Collector->>RSS: GET fetch_url

    alt RSS fetch and ingest succeed
      RSS-->>Collector: RSS feed
      Note over Collector: Normalize RSS items and generate dedupe_key.
      Collector->>Article: POST /internal/ingest
      Article->>DB: bulk upsert articles
      Article-->>Collector: inserted_count, duplicated_count
      Collector->>Article: POST /internal/fetch-jobs/{id}/finish success
      Article->>DB: update fetch_jobs and source status
      Article-->>Collector: 204 No Content
    else RSS fetch or ingest fails
      Collector->>Article: POST /internal/fetch-jobs/{id}/finish failed
      Article->>DB: update fetch_jobs and source status
      Article-->>Collector: 204 No Content
      Note over Collector: Current implementation returns partial results<br/>and stops on the first error.
    end
  end

  Collector-->>Caller: 202 Accepted or error response
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
