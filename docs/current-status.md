# Current Status

## Purpose

このドキュメントは、次のセッションで最初に見るべき「現在地」を短く固定するためのもの。

## Current Phase

- 現在は Phase 1 完了後、Phase 2 着手前

## What Is Already Implemented

- モノレポ構成
- `article-service` の最小 API
  - 記事一覧
  - 記事詳細
  - 検索
  - collector 向け ingest API
- `api-gateway` の最小プロキシ
- `collector-service` の最小 RSS 収集フロー
- `frontend` の記事一覧 / 詳細 / 検索画面
- MySQL 初期スキーマ
- Docker Compose 起動基盤
- ルール / 要件 / ガイドライン / ADR のドキュメント群
- GitHub Issue / PR テンプレート

## Verified State

- Go サービスの `go test ./...` は通過済み
- frontend の `npm run build` は通過済み
- `docker compose config -q` は通過済み

## Highest Priority Next

1. ソース管理 API と UI
2. 取得ジョブ履歴と失敗状況の可視化
3. 既読 / お気に入り管理
4. CSV 出力

## Open GitHub Issues

- `#1` `[Phase 2] ソース管理 API と UI の実装`
- `#2` `[Phase 2] 取得ジョブ履歴と失敗状況の可視化`
- `#3` `[Phase 2] 既読 / お気に入り管理の実装`
- `#4` `[Phase 2] 記事検索結果の CSV 出力実装`
- `#8` `[Phase 3] notification-service と RabbitMQ 連携の実装`
- `#5` `[Phase 4] Compose 整備、GitHub Actions、OpenAPI 保守`
- `#7` `[Phase 5] kind デプロイ、Helm Chart、Argo CD の整備`
- `#6` `[Phase 6] Proxmox クラスタ移行、永続化、監視拡張`

## Known Gaps

- ソース管理はまだ collector の静的設定依存
- 認証は未実装
- 既読 / お気に入りは未実装
- CSV 出力は未実装
- Notification と RabbitMQ は未実装
- kind / Helm / Argo CD は未着手

## Update Rules

以下のタイミングで更新する。

- フェーズが進んだとき
- 優先タスクが変わったとき
- 大きな機能が完了したとき
- 次セッションの入口として重要な状況が変わったとき
