# Backlog Priority

## Purpose

このドキュメントは、Issue が増えても「今どれを先にやるか」を固定するためのもの。

## Priority Order

### Now

1. `#28` Phase 4.5 テスト戦略の実装と回帰防止基盤の強化
2. `#40` Milestone A: 品質ゲート再整理を完了する
3. `#41` Milestone B: 主要境界の自動検証を完了する
4. `#42` Milestone C: テスト保守性と可視化を仕上げる
5. `#43` `make verify` に frontend unit/component test を統合する
6. `#44` CI を `fast` / `integration` / `smoke` の 3 層に再編する
7. `#45` article-service の integration test を CI 常設ジョブに載せる
8. `#46` article-service handler の HTTP integration test を拡充する
9. `#47` collector -> article-service の契約テストを追加する
10. `#48` notification-service の DB integration test を追加する
11. `#49` API の error case テストを強化する
12. `#50` frontend テスト共通基盤とテストデータ戦略を整理する
13. `#51` Playwright E2E の seed 安定化と可視化を進める

### Next

14. `#7` kind デプロイ、Helm Chart、Argo CD の整備
15. 認証導入
16. collector の収集戦略拡張

### Later

17. `#6` Proxmox クラスタ移行、永続化、監視拡張
18. 監視基盤の詳細設計
19. バックアップ / リストア運用の整理

## Priority Rules

- Phase 順を基本にする
- article-service の価値を高めるものを優先する
- 記事閲覧体験に直結するものを優先する
- テストで既存価値を守るための整備は、次フェーズの基盤投資より先に実施してよい
- 監視や本番運用系は、アプリの中核機能が揃ってから進める

## Reordering Rules

優先順を変えるのは以下の場合。

- ブロッカーが出た
- 学習目的が変わった
- 運用上の痛みが実装優先度を上回った
- 依存関係上、先にやるべき基盤タスクが出た

優先順を変えたら `docs/current-status.md` も更新する。
