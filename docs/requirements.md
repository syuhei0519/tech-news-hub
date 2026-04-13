# Requirements

## Overview

個人開発兼学習用として、DevOps / インフラ / クラウド / SRE / CI/CD 関連の情報を継続収集する Web アプリを作る。

### Main Objectives

- 情報収集を効率化すること
- マイクロサービス、Kubernetes、CI/CD、運用監視を学ぶこと

## Assumed Environment

- ローカル開発は Docker Compose
- Kubernetes 検証は kind
- 本運用は Proxmox 上のミニ PC に構築する Kubernetes クラスタ
- DB は MySQL
- 初期は単一ユーザー前提
- 生成 AI は使わず、検索、フィルタ、CSV 出力を中心価値とする

## Functional Requirements

- 外部技術情報ソースを定期収集できること
- 記事や更新情報を保存できること
- 記事一覧、詳細を閲覧できること
- 検索、フィルタ、ソート、ページネーションができること
- 既読、お気に入り管理ができること
- ソースの登録、編集、有効 / 無効切り替えができること
- CSV 出力ができること
- 取得失敗やジョブ状態を確認できること
- 新着や失敗を通知できること

## Non-Functional Requirements

- マイクロサービスとして責務分離されていること
- サービス単位で障害影響を局所化できること
- article 系機能は他サービス停止時も継続利用できること
- 重複記事を防止できること
- 取得失敗時に再試行できること
- ログとメトリクスで問題を追跡できること
- kind 上で動作確認できること
- Proxmox 実クラスタへ移行しやすいこと
- Helm / GitOps で継続運用しやすいこと
- ミニ PC 32GB メモリでも無理なく運用できること

## Learning Requirements

- API Gateway を含むマイクロサービス構成を経験できること
- 非同期処理を導入できること
- Kubernetes でのデプロイと運用を学べること
- Helm / Argo CD による GitOps を学べること
- 監視基盤の基本を学べること
- kind から実クラスタへの移行を経験できること

## Collection Targets

- Google Cloud Release Notes
- Kubernetes Blog
- Kubernetes Release 情報
- GitHub Changelog
- 技術ブログ
- 障害報告 / ポストモーテム記事
- 企業の技術ブログ
- 将来的には企業の IR / 決算資料

## Service Composition

- frontend
- api-gateway
- collector-service
- article-service
- notification-service

## Supporting Components

- MySQL
- Redis
- RabbitMQ
- ingress-nginx
- Prometheus
- Grafana
- Loki

## Data Models

### Article

- id
- title
- url
- source_id
- published_at
- fetched_at
- excerpt
- category
- tags
- is_read
- is_favorite
- dedupe_key
- created_at
- updated_at

### Source

- id
- name
- type
- fetch_url
- fetch_method
- interval_minutes
- default_category
- is_enabled
- last_fetched_at
- last_fetch_status
- last_error_message
- created_at
- updated_at

### Fetch Job

- id
- source_id
- started_at
- finished_at
- status
- fetched_count
- inserted_count
- duplicated_count
- error_message

## Technical Stack

### Frontend

- React
- TypeScript
- Vite
- React Router
- TanStack Query
- Zustand
- Tailwind CSS
- Axios
- React Hook Form
- Zod

### Backend

- Go
- Gin

### Data Access

- MySQL
- sqlc + go-sql-driver/mysql or GORM

この題材では、初期は GORM も現実的だが、堅めに寄せるなら sqlc + 生 SQL の選択肢を持つ。

### Async / Supporting

- Redis
- RabbitMQ
- JWT
- bcrypt
- OpenAPI
- Swagger UI
- GitHub Actions
- Helm
- Argo CD

## Delivery Phases

### Phase 1

- article-service
- collector-service
- api-gateway
- frontend
- MySQL 接続
- 記事一覧 / 詳細 / 検索

### Phase 2

- ソース管理
- 既読 / お気に入り
- CSV 出力
- ジョブ履歴

### Phase 3

- notification-service
- RabbitMQ
- 日次 / 週次ダイジェスト
- 取得失敗通知

### Phase 4

- Docker Compose 整備
- GitHub Actions
- OpenAPI 整備

### Phase 5

- kind デプロイ
- Helm chart 作成
- Argo CD

### Phase 6

- Proxmox 実クラスタ
- 永続化
- 監視強化
- 運用改善
