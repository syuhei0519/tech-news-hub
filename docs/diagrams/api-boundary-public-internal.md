# 公開 API / 内部 API 責務境界図

この図は、frontend の主要 route がどの公開 API group に依存し、その先で `api-gateway`、`article-service`、`notification-service`、`collector-service` がどこで責務分離されているかを 1 枚で把握するためのものです。
通常閲覧導線と collector 導線を分けて示し、`/api/v1` と `/internal` の境界、CSV export の例外経路、OpenAPI 更新や service 境界逸脱の影響点を追いやすくします。

```mermaid
flowchart LR
  classDef route fill:#eef2ff,stroke:#4f46e5,color:#111827,stroke-width:1px;
  classDef public fill:#ecfeff,stroke:#0891b2,color:#111827,stroke-width:1px;
  classDef internal fill:#fff7ed,stroke:#ea580c,color:#111827,stroke-width:1px;
  classDef collector fill:#f8fafc,stroke:#64748b,color:#111827,stroke-width:1px;

  subgraph FE["Frontend Routes"]
    direction TB
    RHome["/<br/>article list / filters / CSV trigger"]
    RArticle["/articles/:id"]
    RSources["/sources"]
    RSourceDetail["/sources/:id"]
    RNotifications["/notifications"]
  end

  subgraph GW["Public API (Gateway /api/v1, mostly thin proxy)"]
    direction TB
    GArticles["articles group<br/>GET /api/v1/articles<br/>GET /api/v1/articles/:id<br/>PATCH /api/v1/articles/:id/read-status<br/>PATCH /api/v1/articles/:id/favorite-status"]
    GSources["sources group<br/>GET /api/v1/sources<br/>POST /api/v1/sources<br/>GET /api/v1/sources/:id<br/>PATCH /api/v1/sources/:id<br/>DELETE /api/v1/sources/:id<br/>GET /api/v1/fetch-jobs"]
    GNotifications["notifications group<br/>GET /api/v1/notifications<br/>PATCH /api/v1/notifications/:id/read-status"]
    GExports["CSV export exception<br/>GET /api/v1/exports/articles.csv"]
  end

  subgraph UP["Upstream Public APIs (/api/v1)"]
    direction TB
    AArticles["article-service<br/>articles<br/>GET /api/v1/articles<br/>GET /api/v1/articles/:id<br/>PATCH /api/v1/articles/:id/read-status<br/>PATCH /api/v1/articles/:id/favorite-status"]
    ASources["article-service<br/>sources + fetch-jobs<br/>GET /api/v1/sources<br/>POST /api/v1/sources<br/>GET /api/v1/sources/:id<br/>PATCH /api/v1/sources/:id<br/>DELETE /api/v1/sources/:id<br/>GET /api/v1/fetch-jobs"]
    AExports["article-service<br/>exports<br/>GET /api/v1/articles/export.csv"]
    NNotifications["notification-service<br/>notifications<br/>GET /api/v1/notifications<br/>PATCH /api/v1/notifications/:id/read-status"]
  end

  subgraph IN["Internal APIs (/internal)"]
    direction TB
    AInternal["article-service collector-only group<br/>POST /internal/fetch-jobs/start<br/>POST /internal/ingest<br/>POST /internal/fetch-jobs/:id/finish"]
  end

  subgraph COL["Collector Entry"]
    direction TB
    CEntry["collector-service entry<br/>POST /api/v1/collect/run"]
    CWorker["collector-service<br/>source sync + RSS fetch + normalize"]
  end

  RHome -->|list / search| GArticles
  RHome -->|source filter options| GSources
  RHome -->|CSV download| GExports
  RArticle -->|detail + status update| GArticles
  RSources -->|source list + save| GSources
  RSourceDetail -->|source detail + fetch-jobs| GSources
  RNotifications -->|list + read-status| GNotifications

  GArticles -->|proxy| AArticles
  GSources -->|proxy| ASources
  GNotifications -->|proxy| NNotifications
  GExports -. path + download header adjusted in gateway .-> AExports

  CEntry --> CWorker
  CWorker -->|load enabled sources| ASources
  CWorker -->|start job / ingest / finish| AInternal

  class RHome,RArticle,RSources,RSourceDetail,RNotifications route;
  class GArticles,GSources,GNotifications,GExports,AArticles,ASources,AExports,NNotifications public;
  class AInternal internal;
  class CEntry,CWorker collector;
```

## 補足

- `api-gateway` は通常 `path / query / body / header` を upstream に渡す thin proxy で、CSV export だけ `GET /api/v1/exports/articles.csv -> GET /api/v1/articles/export.csv` と `Content-Disposition` 調整を行う例外です。
- `collector-service` は通常閲覧導線とは別入口で、`GET /api/v1/sources` と `POST /internal/*` を使って `article-service` と連携します。
- `Internal APIs (/internal)` は collector 連携用であり、frontend や `api-gateway` の公開導線には載せていません。

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
