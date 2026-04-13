# Tech Feed Hub

DevOps / インフラ / クラウド / SRE / CI/CD 関連の技術情報を定期収集して閲覧するための学習用モノレポです。

## Repository Layout

- `frontend`: React + TypeScript + Vite のUI
- `services/api-gateway`: フロント向けBFF
- `services/article-service`: 記事管理API
- `services/collector-service`: 外部ソース収集API / ジョブ
- `services/notification-service`: 通知サービス雛形
- `deployments/compose`: Docker Compose 向け設定
- `deployments/k8s`: kind / Helm / Argo CD 向け土台
- `docs`: API設計・フェーズ計画

## Phase 1

Phase 1 では以下を実装対象とします。

- article-service
- collector-service の最小収集フロー
- api-gateway
- frontend
- MySQL 接続
- 記事一覧 / 詳細 / 検索

## Planned Repository Structure

```text
.
├── frontend
├── services
│   ├── api-gateway
│   ├── article-service
│   ├── collector-service
│   └── notification-service
├── deployments
│   ├── compose
│   ├── k8s
│   └── docker
├── docs
└── scripts
```
