# Delivery Phases

## Phase 1

- article-service
- collector-service
- api-gateway
- frontend
- MySQL
- 記事一覧 / 詳細 / 検索

### Files / Responsibilities

- `services/article-service`: 記事保存、一覧/詳細/検索、collector向け内部登録API
- `services/collector-service`: RSS収集、正規化、article-service への投入
- `services/api-gateway`: フロント向け `/api/v1/articles` 公開
- `frontend`: 記事一覧・詳細・検索UI
- `deployments/compose/mysql/init`: 初期スキーマ
- `docs/openapi/api-gateway.yaml`: 公開API定義

## Phase 2

- ソース管理
- 既読 / お気に入り
- CSV出力
- ジョブ履歴

## Phase 3

- notification-service
- RabbitMQ
- ダイジェスト
- 取得失敗通知

## Phase 4

- Docker Compose 改善
- GitHub Actions
- OpenAPI 拡充

## Phase 5

- kind
- Helm
- Argo CD

## Phase 6

- Proxmox 実クラスタ
- 永続化
- 監視強化
