# Current Status

## Purpose

このドキュメントは、次のセッションで最初に見るべき「現在地」を短く固定するためのもの。

## Current Phase

- 現在は Phase 4 の Compose / GitHub Actions / OpenAPI 整備まで実装済み

## Working Snapshot

- 記事閲覧、検索、CSV 出力、既読 / お気に入り更新は実装済み
- ソース管理、source 詳細、取得ジョブ履歴確認は実装済み
- collector-service は source 管理 API と同期しながら RSS 収集と ingest を実行する
- notification-service と frontend の通知一覧 / 既読更新は実装済み
- Docker Compose、GitHub Actions、公開 OpenAPI 追従まで整備済み

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
