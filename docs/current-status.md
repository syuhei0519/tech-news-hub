# Current Status

## Purpose

このドキュメントは、次のセッションで最初に見るべき「現在地」を短く固定するためのもの。

## Current Phase

- 現在は Phase 4.5 を完了し、次は Phase 5 に進む段階
- kind / Helm / Argo CD の整備を次の着手フェーズとする

## Working Snapshot

- 記事閲覧、検索、CSV 出力、既読 / お気に入り更新は実装済み
- ソース管理、source 詳細、取得ジョブ履歴確認は実装済み
- collector-service は source 管理 API と同期しながら RSS 収集と ingest を実行する
- notification-service と frontend の通知一覧 / 既読更新は実装済み
- Docker Compose、GitHub Actions、公開 OpenAPI 追従まで整備済み
- 最小 E2E smoke は導入済みで、compose seed を使った deterministic 実行へ寄せている
- article-service / collector-service / notification-service / frontend の主要境界テストは整備済み
- Phase 4.5 の親子 Issue はクローズ済みで、次優先は Phase 5 の Kubernetes 系整備である

## Verified State

- Go サービスの `go test ./...` は通過済み
- frontend の `npm run build` は通過済み
- `docker compose config -q` は通過済み
- `make verify` は通過済み
- Playwright smoke と CI レーンは導入済みで、seed 安定化と coverage 可視化が次の仕上げ対象

## Highest Priority Next

1. `#7` kind / Helm / Argo CD の整備
2. 認証導入
3. collector の収集戦略拡張
4. `#6` Proxmox クラスタ移行、永続化、監視拡張
5. 監視基盤の詳細設計

## Open GitHub Issues

- `#7` `[Phase 5] kind デプロイ、Helm Chart、Argo CD の整備`
- `#6` `[Phase 6] Proxmox クラスタ移行、永続化、監視拡張`

## Known Gaps

- 認証は未実装
- kind / Helm / Argo CD は未着手で、これが次の主対象
- frontend の component test と最小 E2E smoke は導入済みだが、共通 helper 整理と可視化は継続整備中
- coverage artifact は導入済みで、OpenAPI 差分検知は将来の強化候補

## Update Rules

以下のタイミングで更新する。

- フェーズが進んだとき
- 優先タスクが変わったとき
- 大きな機能が完了したとき
- 次セッションの入口として重要な状況が変わったとき
