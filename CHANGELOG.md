# Changelog

このプロジェクトの重要な変更を記録する。

## Unreleased

### Changed

- `AGENT.md` と `README.md` の現在地を Phase 1 完了後の状態に同期
- `docs/current-status.md` と `docs/known-issues.md` をソース管理実装後の状態に更新

### Added

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
