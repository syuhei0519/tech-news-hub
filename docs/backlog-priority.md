# Backlog Priority

## Purpose

このドキュメントは、Issue が増えても「今どれを先にやるか」を固定するためのもの。

## Priority Order

### Now

1. `#7` kind デプロイ、Helm Chart、Argo CD の整備
2. 認証導入
3. collector の収集戦略拡張

### Next

4. `#6` Proxmox クラスタ移行、永続化、監視拡張
5. 監視基盤の詳細設計
6. バックアップ / リストア運用の整理

### Later

7. ログ / メトリクス運用の詳細化
8. 本番運用 runbook の拡張

## Priority Rules

- Phase 順を基本にする
- article-service の価値を高めるものを優先する
- 記事閲覧体験に直結するものを優先する
- テストで既存価値を守るための整備は、次フェーズの基盤投資より先に実施してよい
- ただし closed issue を優先順位の Now に残さない
- 監視や本番運用系は、アプリの中核機能が揃ってから進める

## Reordering Rules

優先順を変えるのは以下の場合。

- ブロッカーが出た
- 学習目的が変わった
- 運用上の痛みが実装優先度を上回った
- 依存関係上、先にやるべき基盤タスクが出た

優先順を変えたら `docs/current-status.md` も更新する。
