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
- 最小 E2E smoke は導入済みで、compose seed を使った deterministic 実行へ寄せている
- Phase 4.5 として、テスト戦略の実装と回帰防止基盤の強化を最優先で進める

## Verified State

- Go サービスの `go test ./...` は通過済み
- frontend の `npm run build` は通過済み
- `docker compose config -q` は通過済み
- `make verify` は通過済み
- Playwright smoke と CI レーンは導入済みで、seed 安定化と coverage 可視化が次の仕上げ対象

## Highest Priority Next

1. Phase 4.5 のテスト戦略実装と回帰防止基盤の強化
2. Milestone A: 品質ゲート再整理
3. Milestone B: 主要境界の自動検証
4. Milestone C: テスト保守性と可視化
5. kind / Helm / Argo CD の整備

## Open GitHub Issues

- `#28` `[Phase 4.5]` テスト戦略の実装と回帰防止基盤の強化
- `#40` `[Phase 4.5][Milestone A]` 品質ゲート再整理を完了する
- `#41` `[Phase 4.5][Milestone B]` 主要境界の自動検証を完了する
- `#42` `[Phase 4.5][Milestone C]` テスト保守性と可視化を仕上げる
- `#43` `[Phase 4.5]` `make verify` に frontend unit/component test を統合する
- `#44` `[Phase 4.5]` CI を `fast` / `integration` / `smoke` の 3 層に再編する
- `#45` `[Phase 4.5]` article-service の integration test を CI 常設ジョブに載せる
- `#46` `[Phase 4.5]` article-service handler の HTTP integration test を拡充する
- `#47` `[Phase 4.5]` collector -> article-service の契約テストを追加する
- `#48` `[Phase 4.5]` notification-service の DB integration test を追加する
- `#49` `[Phase 4.5]` API の error case テストを強化する
- `#50` `[Phase 4.5]` frontend テスト共通基盤とテストデータ戦略を整理する
- `#51` `[Phase 4.5]` Playwright E2E の seed 安定化と可視化を進める
- `#7` `[Phase 5] kind デプロイ、Helm Chart、Argo CD の整備`
- `#6` `[Phase 6] Proxmox クラスタ移行、永続化、監視拡張`

## Known Gaps

- 認証は未実装
- kind / Helm / Argo CD は未着手で、Phase 4.5 完了後に着手する
- frontend の component test と最小 E2E smoke は導入済みだが、共通 helper 整理と可視化は継続整備中
- repository / DB 境界と service 間契約の自動検証は未整備

## Update Rules

以下のタイミングで更新する。

- フェーズが進んだとき
- 優先タスクが変わったとき
- 大きな機能が完了したとき
- 次セッションの入口として重要な状況が変わったとき
