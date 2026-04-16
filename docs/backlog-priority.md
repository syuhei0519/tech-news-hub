# Backlog Priority

## Purpose

このドキュメントは、Issue が増えても「今どれを先にやるか」を固定するためのもの。

## Priority Order

### Now

1. `#7` kind デプロイ、Helm Chart、Argo CD の整備
2. `#6` Proxmox クラスタ移行、永続化、監視拡張

### Next

3. 認証導入
4. 監視基盤の詳細設計
5. collector の収集戦略拡張

### Later

6. バックアップ / リストア運用の整理

## Priority Rules

- Phase 順を基本にする
- article-service の価値を高めるものを優先する
- 記事閲覧体験に直結するものを優先する
- 監視や本番運用系は、アプリの中核機能が揃ってから進める

## Reordering Rules

優先順を変えるのは以下の場合。

- ブロッカーが出た
- 学習目的が変わった
- 運用上の痛みが実装優先度を上回った
- 依存関係上、先にやるべき基盤タスクが出た

優先順を変えたら `docs/current-status.md` も更新する。
