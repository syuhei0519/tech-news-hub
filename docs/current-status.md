# Current Status

## Purpose

このドキュメントは、次のセッションで最初に見るべき「現在地」を短く固定するためのもの。

## Current Phase

- 現在は Phase 3 の notification-service / RabbitMQ 連携まで実装済み

## What Is Already Implemented

- モノレポ構成
- `article-service` の最小 API
  - 記事一覧
  - 記事詳細
  - 検索
  - collector 向け ingest API
- `article-service` のソース管理 API
  - ソース一覧
  - ソース詳細
  - ソース作成
  - ソース更新
  - ソース削除
- `api-gateway` の最小プロキシ
- `api-gateway` 経由のソース管理 API 公開
- `collector-service` の最小 RSS 収集フロー
- `frontend` の記事一覧 / 詳細 / 検索画面
- `frontend` の記事検索結果 CSV ダウンロード
- `frontend` のソース一覧 / 作成 / 編集 / 有効無効切り替え画面
- `frontend` の source 詳細と取得ジョブ履歴確認画面
- `notification-service` の通知一覧 / 既読更新 API
- `frontend` の通知一覧画面
- RabbitMQ 経由の新着通知 / 取得失敗通知フロー
- MySQL 初期スキーマ
- Docker Compose 起動基盤
- ルール / 要件 / ガイドライン / ADR のドキュメント群
- GitHub Issue / PR テンプレート
- GitHub Actions の最小 CI
- collector-service / api-gateway の最小ユニットテスト

## Verified State

- Go サービスの `go test ./...` は通過済み
- frontend の `npm run build` は通過済み
- `docker compose config -q` は通過済み
- `make verify` は通過済み

## Highest Priority Next

1. collector-service の動的ソース同期
2. Compose 整備、GitHub Actions、OpenAPI 保守
3. kind / Helm / Argo CD の整備
4. Proxmox クラスタ移行、永続化、監視拡張

## Open GitHub Issues

- `#5` `[Phase 4] Compose 整備、GitHub Actions、OpenAPI 保守`
- `#7` `[Phase 5] kind デプロイ、Helm Chart、Argo CD の整備`
- `#6` `[Phase 6] Proxmox クラスタ移行、永続化、監視拡張`

## Known Gaps

- ソース管理 UI/API とジョブ履歴 UI/API は実装済みだが、collector-service はまだ静的設定依存
- 認証は未実装
- kind / Helm / Argo CD は未着手

## Update Rules

以下のタイミングで更新する。

- フェーズが進んだとき
- 優先タスクが変わったとき
- 大きな機能が完了したとき
- 次セッションの入口として重要な状況が変わったとき
