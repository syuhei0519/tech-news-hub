# Tech Feed Hub

DevOps / インフラ / クラウド / SRE / CI/CD 関連の技術情報を定期収集して閲覧するための学習用モノレポです。

## Repository Layout

- `frontend`: React + TypeScript + Vite のUI
- `services/api-gateway`: フロント向けBFF
- `services/article-service`: 記事管理API
- `services/collector-service`: 外部ソース収集API / ジョブ
- `services/notification-service`: 通知生成・一覧 API サービス
- `deployments/compose`: Docker Compose 向け設定
- `deployments/k8s`: kind / Helm / Argo CD 向け土台
- `docs`: API設計・フェーズ計画

## Architecture And Diagrams

- runtime 構成、記事収集、通知生成、CI / テストレーンの図は [図ドキュメント](docs/diagrams/README.md) から参照できます
- 図は現行 repo 実装を根拠にしており、`tech-news-hub/` 配下や未実装の将来構成は含めていません

## Current Status

Phase 1 の記事閲覧フローに加えて、Phase 2 の主要機能、Phase 3 の通知経路、Phase 4 の基盤整備、Phase 4.5 のテスト強化まで実装済みです。次段は Phase 5 の kind / Helm / Argo CD 整備です。

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

- Phase 4.5 の品質ゲート、integration test、component test、最小 E2E smoke まで整備済み
- 次優先は Phase 5 の kind / Helm / Argo CD 整備
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
- `docs/agent-playbooks.md`: エージェント向けの最小読込プレイブック
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
- `docs/openapi/api-gateway.summary.md`: 公開 API の要約導線
- `CHANGELOG.md`: 重要変更履歴

## Standard Commands

- `make up`: 全サービス起動
- `make down`: volume を残して停止
- `make reset`: volume を削除して初期化
- `make dev-up`: frontend HMR と Go auto-reload 付きで起動
- `make dev-down`: dev compose を停止
- `make dev-logs`: dev compose のログ追跡
- `make dev-ps`: dev compose の状態確認
- `make build`: frontend build
- `make test`: backend test
- `make test-frontend`: frontend unit/component test
- `make test-frontend-coverage`: frontend unit/component test + coverage
- `make test-article-integration`: article-service の MySQL integration test
- `make test-notification-integration`: notification-service の MySQL integration test
- `make verify`: backend test + frontend test + build + compose config check
- `make e2e-up` / `make e2e-seed` / `make test-e2e` / `make e2e-down`: Playwright E2E smoke

## Testing Overview

- 日常の基本確認ルートは `make verify`
- DB 境界は `integration` レーンで article-service / notification-service の MySQL integration test を常時確認
- 主要ユーザーフローは `smoke` レーンで Playwright E2E を relevant changes 時に確認
- frontend coverage は `coverage` レーンで artifact 化する

テストレイヤの役割:

- Unit / Component: 日常変更で壊れやすいロジックと画面導線を高速に確認する
- Integration: repository / DB / handler 境界を確認する
- E2E smoke: 記事導線と通知導線の最小フローだけを確認する

詳細な基準は `docs/test-policy.md` を参照してください。

補足:

- 通常のコード変更では CI を `verify` / `integration` / `coverage` / `smoke` のレーンで実行します
- `verify` は `make verify` を実行する高速レーンです
- `integration` は article-service と notification-service の MySQL integration test を実行します
- `coverage` は frontend coverage artifact を生成します
- `smoke` は Playwright E2E を relevant changes 時に実行します
- docs-only 変更では、GitHub Actions は `verify` を軽量成功で完了させつつ重い検証を省略します
- collector-service は `ARTICLE_SERVICE_URL` 経由で source 一覧を取得し、`is_enabled=true` の source を収集します
- ローカルで即時反映が必要な場合は `docker-compose.dev.yml` を重ねる `make dev-up` を使います
