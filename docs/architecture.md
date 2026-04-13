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

## Kubernetes Readiness

- コンテナ設定は環境変数で外出し
- CronJob 化しやすいよう collector の収集処理を HTTP/CLI 両対応にしやすい構成にする
- `deployments/k8s` に Helm values を増やせる前提でディレクトリを確保する
