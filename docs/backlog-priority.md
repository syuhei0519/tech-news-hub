# Backlog Priority

## Purpose

このドキュメントは、Issue が増えても「今どれを先にやるか」を固定するためのもの。

## Priority Order

### Now

1. collector-service の動的ソース同期
2. `#8` notification-service と RabbitMQ 連携の実装
3. `#5` Compose 整備、GitHub Actions、OpenAPI 保守

### Next

4. `#7` kind デプロイ、Helm Chart、Argo CD の整備
5. `#6` Proxmox クラスタ移行、永続化、監視拡張

### Later

6. 認証導入
7. 監視基盤の詳細設計

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
