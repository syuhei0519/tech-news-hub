# Known Issues

## Purpose

このドキュメントは、既知の制約や未実装事項を明文化し、新規セッションで誤解を減らすためのもの。

## Current Known Issues

### Collector Source Configuration Is Still Static

- 現在の collector-service は環境変数ベースのソース定義を使用している
- ソース管理 API / UI はあるが、collector-service とはまだ接続されていない
- アプリ上の source 作成・編集・有効無効切り替えは収集対象へ自動反映されない
- Phase 2 の後続対応で collector 側連携が必要

### Authentication Is Not Implemented

- api-gateway に JWT の導入ポイントはあるが、実装は未着手
- 初期は単一ユーザー前提のため後ろ倒し

### CSV Export Is Not Implemented Yet

- 要件にはあるが、現状は記事一覧 / 詳細 / 検索まで
- Phase 2 の `#4` で対応予定

### Notification Flow Is Not Implemented Yet

- notification-service は health endpoint のみ
- RabbitMQ も未接続
- Phase 3 の `#8` で対応予定

### Kubernetes Delivery Is Not Implemented Yet

- `deployments/k8s` はプレースホルダのみ
- kind / Helm / Argo CD は未着手

## Usage Rules

- 既知制約を見つけたらここに追加する
- 既知課題が解消されたら削除または更新する
- PR で「未対応だが既知」のものはここに記載する
