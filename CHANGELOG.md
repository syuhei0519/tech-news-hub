# Changelog

このプロジェクトの重要な変更を記録する。

## Unreleased

### Changed

- 進捗系ドキュメントを Phase 2 の主要機能実装後、Phase 3 着手前の状態に同期
- `docs/current-status.md` の Open GitHub Issues を現行の open issue に更新
- `docs/backlog-priority.md` の優先順位を現状の残課題と open issue に合わせて更新
- 記事一覧 API と CSV エクスポートで `source_id`, `from`, `to` を含む共通フィルタを利用
- `api-gateway` の CSV API 方針を `GET /api/v1/exports/articles.csv` に統一

### Added

- article-service の記事 CSV エクスポート API
- api-gateway 経由の CSV ダウンロード API
- frontend の source / 公開日範囲フィルタと CSV ダウンロード導線
- article-service のソース管理 CRUD API
- api-gateway 経由のソース管理 API 公開
- frontend のソース一覧 / 作成 / 編集 / 有効無効切り替え画面
- source API を含む OpenAPI 定義
- ルール運用のためのドキュメント群
- API / DB / Kubernetes ガイドライン
- Definition of Done
- Runbook
- Test policy
- Naming conventions
- Release and migration policy
- ADR テンプレートと初回 ADR
- GitHub Issue / PR テンプレート
- Current status / backlog priority / known issues / glossary
- Makefile の `build`, `test`, `verify` タスク
- GitHub Actions の最小 CI workflow
- collector-service / api-gateway の最小テスト

## 0.1.0

### Added

- 初期モノレポ構成
- article-service の最小 API
- api-gateway の最小プロキシ
- collector-service の最小 RSS 収集フロー
- notification-service の雛形
- frontend の一覧 / 詳細 / 検索画面
- Docker Compose によるローカル起動基盤
