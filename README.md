# Tech Feed Hub

DevOps / インフラ / クラウド / SRE / CI/CD 関連の技術情報を定期収集して閲覧するための学習用モノレポです。

## Repository Layout

- `frontend`: React + TypeScript + Vite のUI
- `services/api-gateway`: フロント向けBFF
- `services/article-service`: 記事管理API
- `services/collector-service`: 外部ソース収集API / ジョブ
- `services/notification-service`: 通知サービス雛形
- `deployments/compose`: Docker Compose 向け設定
- `deployments/k8s`: kind / Helm / Argo CD 向け土台
- `docs`: API設計・フェーズ計画

## Current Status

Phase 1 の記事閲覧フローに加えて、Phase 2 の主要機能、Phase 3 の通知経路、Phase 4 の基盤整備を実装済みです。

- article-service / collector-service / api-gateway / frontend
- MySQL 接続
- 記事一覧 / 詳細 / 検索
- 既読 / お気に入り管理
- ソース管理
- 記事検索結果の CSV 出力
- 取得ジョブ履歴の表示
- RabbitMQ 経由の通知生成
- 通知一覧 / 既読更新
- Docker Compose のローカル運用整備
- GitHub Actions の最小 CI
- OpenAPI の公開 API 追従

現在地:

- Phase 4 の Compose / CI / OpenAPI 整備まで実装済み
- collector-service は article-service の source 管理 API から実行時に収集対象を同期
- 検証状態の詳細は `docs/current-status.md` を参照

## Planned Repository Structure

```text
.
├── frontend
├── services
│   ├── api-gateway
│   ├── article-service
│   ├── collector-service
│   └── notification-service
├── deployments
│   ├── compose
│   ├── k8s
│   └── docker
├── docs
└── scripts
```

## Rule Documents

- `AGENT.md`: 開発者向けの入口
- `RULES.md`: プロジェクト全体ルール
- `docs/requirements.md`: 要件整理
- `docs/rules-operations.md`: ルール運用方法
- `docs/dod.md`: 完了条件
- `docs/runbook.md`: 開発 / 切り分け手順
- `docs/test-policy.md`: テスト方針
- `docs/naming-conventions.md`: 命名規則
- `docs/release-migration-policy.md`: 変更と移行の方針
- `docs/branch-strategy.md`: ブランチ運用方針
- `docs/adr/`: 重要な設計判断の記録
- `docs/current-status.md`: 現在地
- `docs/backlog-priority.md`: 優先順位
- `docs/known-issues.md`: 既知課題
- `docs/glossary.md`: 用語集
- `CHANGELOG.md`: 重要変更履歴

## Standard Commands

- `make up`: 全サービス起動
- `make down`: volume を残して停止
- `make reset`: volume を削除して初期化
- `make build`: frontend build
- `make test`: backend test
- `make verify`: test + build + compose config check

補足:

- 通常のコード変更では CI も `make verify` 相当を実行します
- docs-only 変更では、GitHub Actions は required check を維持しつつ重い検証を省略します
- collector-service は `ARTICLE_SERVICE_URL` 経由で source 一覧を取得し、`is_enabled=true` の source を収集します
