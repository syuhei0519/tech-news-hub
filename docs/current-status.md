# Current Status

## Purpose

このドキュメントは、次のセッションで最初に見るべき「現在地」を短く固定するためのもの。

## Current Phase

- 現在は Phase 4 の Compose / GitHub Actions / OpenAPI 整備まで実装済み

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
- `collector-service` の source 管理 API 連携と実行時同期
- `frontend` の記事一覧 / 詳細 / 検索画面
- `frontend` の記事検索結果 CSV ダウンロード
- `frontend` のソース一覧 / 作成 / 編集 / 有効無効切り替え画面
- `frontend` の source 詳細と取得ジョブ履歴確認画面
- `notification-service` の通知一覧 / 既読更新 API
- `frontend` の通知一覧画面
- RabbitMQ 経由の新着通知 / 取得失敗通知フロー
- MySQL 初期スキーマ
- Docker Compose のローカル運用整備
- ルール / 要件 / ガイドライン / ADR のドキュメント群
- GitHub Issue / PR テンプレート
- GitHub Actions の最小 CI
- collector-service / api-gateway の最小ユニットテスト
- OpenAPI の公開 API 追従

## Verified State

- Go サービスの `go test ./...` は通過済み
- frontend の `npm run build` は通過済み
- `docker compose config -q` は通過済み
- `make verify` は通過済み

## Highest Priority Next

1. kind / Helm / Argo CD の整備
2. Proxmox クラスタ移行、永続化、監視拡張
3. 認証導入の設計整理
4. collector の収集戦略拡張

## Open GitHub Issues

- `#7` `[Phase 5] kind デプロイ、Helm Chart、Argo CD の整備`
- `#6` `[Phase 6] Proxmox クラスタ移行、永続化、監視拡張`

## Known Gaps

- 認証は未実装
- kind / Helm / Argo CD は未着手

## Update Rules

以下のタイミングで更新する。

- フェーズが進んだとき
- 優先タスクが変わったとき
- 大きな機能が完了したとき
- 次セッションの入口として重要な状況が変わったとき
