# Backlog Priority

## Purpose

このドキュメントは、Issue が増えても「今どれを先にやるか」を固定するためのもの。

## Priority Order

### Now

1. `#28` Phase 4.5 テスト戦略の実装と回帰防止基盤の強化
2. `#29` article-service の repository / handler / MySQL 結合テストを追加する
3. `#30` frontend のテスト基盤を導入し記事一覧 / 詳細導線を守る
4. `#31` notification / source 管理周辺のテストを強化する
5. `#32` collector のデータ揺れケースと service 間契約テストを追加する
6. `#33` 最小 E2E と CI への組み込み方針を整備する

### Next

7. `#7` kind デプロイ、Helm Chart、Argo CD の整備
8. 認証導入
9. collector の収集戦略拡張

### Later

10. `#6` Proxmox クラスタ移行、永続化、監視拡張
11. 監視基盤の詳細設計
12. バックアップ / リストア運用の整理

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
