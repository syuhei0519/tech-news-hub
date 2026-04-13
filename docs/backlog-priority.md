# Backlog Priority

## Purpose

このドキュメントは、Issue が増えても「今どれを先にやるか」を固定するためのもの。

## Priority Order

### Now

1. `#1` ソース管理 API と UI の実装
2. `#2` 取得ジョブ履歴と失敗状況の可視化
3. `#3` 既読 / お気に入り管理の実装
4. `#4` 記事検索結果の CSV 出力実装

### Next

5. `#8` notification-service と RabbitMQ 連携の実装
6. `#5` Compose 整備、GitHub Actions、OpenAPI 保守

### Later

7. `#7` kind デプロイ、Helm Chart、Argo CD の整備
8. `#6` Proxmox クラスタ移行、永続化、監視拡張

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
