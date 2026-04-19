# Architecture

## Monorepo

- `frontend`: 画面責務
- `api-gateway`: 認証、集約、共通エラー処理
- `article-service`: 記事・ソース・集計の責務
- `collector-service`: 外部取得と正規化の責務
- `notification-service`: 通知生成と配送の責務

## Service Boundaries

- `article-service` が記事閲覧可用性の中心
- `collector-service` 停止時も既存記事閲覧は継続
- `notification-service` 停止時も既存記事閲覧は継続
- Phase 1 は共有 MySQL だが、サービス単位のアクセス責務を分離する

## Async Notifications

- RabbitMQ を topic exchange として使い、`article.ingested` と `collector.fetch.failed` を流す
- `article-service` は ingest 成功時に新着通知イベントを publish する
- `collector-service` は取得失敗時に失敗通知イベントを publish する
- `notification-service` はイベントを購読し、通知レコードを生成して公開 API で返す
- RabbitMQ または `notification-service` の障害は記事閲覧 API の可用性より優先しない

## Related Diagrams

- [システム全体構成図](diagrams/system-overview.md)
- [公開 API / 内部 API 責務境界図](diagrams/api-boundary-public-internal.md)
- [データモデル / 所有責務 ER 図](diagrams/data-model-ownership-er.md)
- [記事収集フロー図](diagrams/article-collection-flow.md)
- [通知生成フロー図](diagrams/notification-generation-flow.md)

## Kubernetes Readiness

- コンテナ設定は環境変数で外出し
- CronJob 化しやすいよう collector の収集処理を HTTP/CLI 両対応にしやすい構成にする
- `deployments/k8s` に Helm values を増やせる前提でディレクトリを確保する
