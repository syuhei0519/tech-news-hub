# Current Status

## Purpose

このドキュメントは、次のセッションで最初に見るべき「現在地」を短く固定するためのもの。

## Current Phase

- 現在は Phase 4 の Compose / GitHub Actions / OpenAPI 整備まで実装済み
- 次の着手フェーズとして、Phase 5 より前に Phase 4.5 のテスト強化を差し込む

## Working Snapshot

- 記事閲覧、検索、CSV 出力、既読 / お気に入り更新は実装済み
- ソース管理、source 詳細、取得ジョブ履歴確認は実装済み
- collector-service は source 管理 API と同期しながら RSS 収集と ingest を実行する
- notification-service と frontend の通知一覧 / 既読更新は実装済み
- Docker Compose、GitHub Actions、公開 OpenAPI 追従まで整備済み
- 最小 E2E は未導入で、component test までが自動化の中心になっている
- Phase 4.5 として、テスト戦略の実装と回帰防止基盤の強化を最優先で進める

## Verified State

- Go サービスの `go test ./...` は通過済み
- frontend の `npm run build` は通過済み
- `docker compose config -q` は通過済み
- `make verify` は通過済み
- E2E は runner と CI レーンをこれから分離導入する段階

## Highest Priority Next

1. Phase 4.5 のテスト戦略実装と回帰防止基盤の強化
2. article-service の repository / handler / MySQL 結合テスト追加
3. frontend の記事一覧 / 詳細導線テスト基盤追加
4. notification / source 管理周辺のテスト強化
5. collector のデータ揺れケースと service 間契約テスト追加
6. 最小 E2E と CI への組み込み
7. kind / Helm / Argo CD の整備

## Open GitHub Issues

- `#28` `[Phase 4.5]` テスト戦略の実装と回帰防止基盤の強化
- `#29` `[Phase 4.5]` article-service の repository / handler / MySQL 結合テストを追加する
- `#30` `[Phase 4.5]` frontend のテスト基盤を導入し記事一覧 / 詳細導線を守る
- `#31` `[Phase 4.5]` notification / source 管理周辺のテストを強化する
- `#32` `[Phase 4.5]` collector のデータ揺れケースと service 間契約テストを追加する
- `#33` `[Phase 4.5]` 最小 E2E と CI への組み込み方針を整備する
- `#7` `[Phase 5] kind デプロイ、Helm Chart、Argo CD の整備`
- `#6` `[Phase 6] Proxmox クラスタ移行、永続化、監視拡張`

## Known Gaps

- 認証は未実装
- kind / Helm / Argo CD は未着手
- frontend の component test は導入済みだが、実ブラウザ E2E は未導入
- repository / DB 境界と service 間契約の自動検証は未整備

## Update Rules

以下のタイミングで更新する。

- フェーズが進んだとき
- 優先タスクが変わったとき
- 大きな機能が完了したとき
- 次セッションの入口として重要な状況が変わったとき
