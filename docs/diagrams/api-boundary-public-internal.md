# 公開 API / 内部 API 責務境界図

この図は、frontend の主要 route がどの公開 API group に依存し、その先で `api-gateway`、`article-service`、`notification-service`、`collector-service` がどこで責務分離されているかを 1 枚で把握するためのものです。
通常閲覧導線と collector 導線を分けて示し、`/api/v1` と `/internal` の境界、CSV export の例外経路、OpenAPI 更新や service 境界逸脱の影響点を追いやすくします。

```mermaid
flowchart LR
  classDef route fill:#eef2ff,stroke:#4f46e5,color:#111827,stroke-width:1px;
  classDef public fill:#ecfeff,stroke:#0891b2,color:#111827,stroke-width:1px;
  classDef internal fill:#fff7ed,stroke:#ea580c,color:#111827,stroke-width:1px;
  classDef collector fill:#f8fafc,stroke:#64748b,color:#111827,stroke-width:1px;

  subgraph FE["Frontend Routes / 画面 route"]
    direction TB
    RHome["/<br/>記事一覧 / フィルタ / CSV 出力"]
    RArticle["/articles/:id"]
    RSources["/sources"]
    RSourceDetail["/sources/:id"]
    RNotifications["/notifications"]
  end

  subgraph GW["公開 API (Gateway /api/v1, ほぼ thin proxy)"]
    direction TB
    GArticles["記事 API group<br/>GET /api/v1/articles<br/>GET /api/v1/articles/:id<br/>PATCH /api/v1/articles/:id/read-status<br/>PATCH /api/v1/articles/:id/favorite-status"]
    GSources["source API group<br/>GET /api/v1/sources<br/>POST /api/v1/sources<br/>GET /api/v1/sources/:id<br/>PATCH /api/v1/sources/:id<br/>DELETE /api/v1/sources/:id<br/>GET /api/v1/fetch-jobs"]
    GNotifications["通知 API group<br/>GET /api/v1/notifications<br/>PATCH /api/v1/notifications/:id/read-status"]
    GExports["CSV export 例外経路<br/>GET /api/v1/exports/articles.csv"]
  end

  subgraph UP["Upstream 公開 API (/api/v1)"]
    direction TB
    AArticles["article-service<br/>記事 API<br/>GET /api/v1/articles<br/>GET /api/v1/articles/:id<br/>PATCH /api/v1/articles/:id/read-status<br/>PATCH /api/v1/articles/:id/favorite-status"]
    ASources["article-service<br/>source + fetch-jobs API<br/>GET /api/v1/sources<br/>POST /api/v1/sources<br/>GET /api/v1/sources/:id<br/>PATCH /api/v1/sources/:id<br/>DELETE /api/v1/sources/:id<br/>GET /api/v1/fetch-jobs"]
    AExports["article-service<br/>export API<br/>GET /api/v1/articles/export.csv"]
    NNotifications["notification-service<br/>通知 API<br/>GET /api/v1/notifications<br/>PATCH /api/v1/notifications/:id/read-status"]
  end

  subgraph IN["内部 API (/internal)"]
    direction TB
    AInternal["article-service collector 専用 group<br/>POST /internal/fetch-jobs/start<br/>POST /internal/ingest<br/>POST /internal/fetch-jobs/:id/finish"]
  end

  subgraph COL["Collector Entry / 収集入口"]
    direction TB
    CEntry["collector-service 入口<br/>POST /api/v1/collect/run"]
    CWorker["collector-service<br/>source 同期 + RSS 取得 + 正規化"]
  end

  RHome -->|一覧 / 検索| GArticles
  RHome -->|source フィルタ候補| GSources
  RHome -->|CSV 出力| GExports
  RArticle -->|詳細 / 状態更新| GArticles
  RSources -->|source 一覧 / 保存| GSources
  RSourceDetail -->|source 詳細 / fetch-jobs| GSources
  RNotifications -->|一覧 / 既読更新| GNotifications

  GArticles -->|proxy| AArticles
  GSources -->|proxy| ASources
  GNotifications -->|proxy| NNotifications
  GExports -. gateway で path と download header を調整 .-> AExports

  CEntry --> CWorker
  CWorker -->|enabled な source を取得| ASources
  CWorker -->|job 開始 / ingest / job 完了| AInternal

  class RHome,RArticle,RSources,RSourceDetail,RNotifications route;
  class GArticles,GSources,GNotifications,GExports,AArticles,ASources,AExports,NNotifications public;
  class AInternal internal;
  class CEntry,CWorker collector;
```

## 補足

- `api-gateway` は通常 `path / query / body / header` を upstream に渡す thin proxy で、CSV export だけ `GET /api/v1/exports/articles.csv -> GET /api/v1/articles/export.csv` と `Content-Disposition` の調整を行う例外です。
- `collector-service` は通常閲覧導線とは別入口で、`GET /api/v1/sources` と `POST /internal/*` を使って `article-service` と連携します。
- `内部 API (/internal)` は collector 連携用であり、frontend や `api-gateway` の公開導線には載せていません。

## 主な根拠ファイル

- `frontend/src/main.tsx`
- `frontend/src/lib/api.ts`
- `frontend/src/ui/ArticleListPage.tsx`
- `frontend/src/ui/ArticleDetailPage.tsx`
- `frontend/src/ui/SourceManagementPage.tsx`
- `frontend/src/ui/SourceDetailPage.tsx`
- `frontend/src/ui/NotificationListPage.tsx`
- `services/api-gateway/internal/httpapi/routes.go`
- `services/article-service/internal/httpapi/router.go`
- `services/collector-service/internal/httpapi/routes.go`
- `services/collector-service/internal/collector/service.go`
- `services/notification-service/internal/httpapi/router.go`
- `docs/openapi/api-gateway.summary.md`
- `docs/openapi/article-service-internal.yaml`
