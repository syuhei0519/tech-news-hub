# AGENT.md

## Purpose

このリポジトリは、DevOps / インフラ / クラウド / SRE / CI/CD 関連の情報を定期収集し、一覧・検索・管理する学習用 Web アプリ `Tech Feed Hub` のモノレポです。

目的は次の 2 点です。

- 技術情報を継続的に収集・閲覧できること
- マイクロサービス、Kubernetes、GitOps、監視運用を段階的に学べること

生成 AI による要約や自動分類は初期スコープに含めません。

## Source Of Truth

要件の詳細は以下を優先して参照します。

- `RULES.md`
- `docs/requirements.md`
- `docs/rules-operations.md`
- `docs/dod.md`
- `docs/test-policy.md`
- `docs/release-migration-policy.md`
- `docs/branch-strategy.md`
- `docs/current-status.md`
- `docs/backlog-priority.md`
- `docs/known-issues.md`
- `docs/glossary.md`
- `docs/architecture.md`
- `docs/phases.md`

## Repository Layout

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
│   ├── docker
│   └── k8s
├── docs
├── docker-compose.yml
├── Makefile
└── AGENT.md
```

## Current Phase

現在は Phase 4 の Compose / CI / OpenAPI 整備まで実装済みです。

Phase 1 で完了した範囲:

- article-service
- collector-service の最小収集フロー
- api-gateway
- frontend
- MySQL 接続
- 記事一覧 / 詳細 / 検索

Phase 2 で完了した範囲:

- 既読 / お気に入り
- ソース管理
- CSV 出力
- 取得ジョブ履歴

Phase 3 で完了した範囲:

- notification-service
- RabbitMQ 連携
- 新着通知と取得失敗通知

Phase 4 で完了した範囲:

- Docker Compose のローカル運用整備
- GitHub Actions の最小 CI
- OpenAPI の公開 API 追従

現在の残課題:

- kind / Helm / Argo CD の整備
- Proxmox クラスタ移行、永続化、監視拡張

次に優先して進める内容は以下です。

- kind / Helm / Argo CD の整備
- Proxmox クラスタ移行、永続化、監視拡張

## Service Responsibilities

### frontend

- React + TypeScript + Vite
- 記事一覧、詳細、検索 UI
- API Gateway 経由でバックエンドアクセス

### api-gateway

- フロント向け単一入口
- 将来の JWT 認証導入ポイント
- 共通ログ、共通エラー制御
- article-service へのプロキシ

### article-service

- 記事保存
- 記事一覧 / 詳細 / 検索
- collector-service からの内部取り込み API
- ソース / 取得ジョブ履歴の基礎テーブル管理

### collector-service

- RSS 取得
- 外部データ正規化
- dedupe_key 生成
- article-service への投入

### notification-service

- 通知一覧 / 既読更新 API
- RabbitMQ 経由の新着通知 / 取得失敗通知
- 後続フェーズで日次 / 週次ダイジェスト、Slack / メール連携を追加

## Local Development

### Requirements

- Docker
- Docker Compose
- Go 1.24+
- Node.js 22+
- npm

### First Setup

```bash
cp .env.example .env
docker compose config -q
```

### Start All Services

```bash
make up
```

### Stop

```bash
make down
```

volume も含めて初期化したい場合:

```bash
make reset
```

## Verification

Go サービス確認:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
```

frontend 確認:

```bash
cd frontend
npm install
npm run build
```

## Implementation Rules

- 過剰設計より、学習しやすさと継続運用しやすさを優先する
- ただしサービス責務は明確に分離する
- 初期は共有 MySQL を許容するが、コード上は各サービスが自分の責務データのみを扱う
- Docker Compose で全体起動できることを優先する
- kind / 実クラスタへ移しやすいよう、設定は環境変数で切り替える
- OpenAPI を前提に API を追加する
- 破壊的変更を入れる場合は、Compose 起動と既存 API の継続性を確認する
- 処理意図がコードだけでは追いにくい箇所には、明示的な指示がなくても日本語コメントを付ける
- コメントは要点だけを短く書き、コードの逐語説明ではなく判断理由や前提を残す

## API Design Rules

- 公開 API は基本的に `api-gateway` 配下に置く
- サービス間通信用エンドポイントは `internal` プレフィックスで分離する
- 検索、フィルタ、ソート、ページネーションを早い段階で意識する
- CSV 出力は将来 `api-gateway` 側で制御する

## Data Rules

- 記事重複防止には `dedupe_key` を使う
- 日時は UTC ベースで保存する
- タグは初期実装では JSON 配列で保持する
- DB スキーマ変更時は `deployments/compose/mysql/init` と今後のマイグレーション方針の整合を意識する

## Deployment Direction

- 開発初期は Docker Compose
- Kubernetes 検証は kind
- その後 Proxmox 上の実クラスタに移行
- GitOps は Helm + Argo CD を想定

以下は環境差分が出やすいため分離を意識すること。

- Ingress
- StorageClass
- Domain
- Secret 管理

## Monitoring Direction

初期は最小構成です。

- health check
- 構造化ログの導入余地
- 将来的に Prometheus / Grafana / Loki を追加

## GitHub Preparation Notes

- `.env` はコミットしない
- `.env.example` を最新に保つ
- `node_modules`, `dist`, `*.tsbuildinfo` はコミットしない
- ドキュメント更新を伴う設計変更では `docs/` も更新する
- Git のコミットメッセージは日本語で記述する
- Pull Request の題名と本文は、明示的な指示がない限り日本語で記述する
- Pull Request 作成時は `.github/pull_request_template.md` を使い、各項目を実施内容に合わせて埋める

## Documentation Update Rules

- エージェントは作業内容に応じて、必要なドキュメント更新を自分で判断して同じ変更内で実施する
- 実装、設計、手順、現在地、優先順位、既知課題のいずれかが変わった場合、コード変更だけで完了扱いにしない
- API 変更時は OpenAPI と API 関連 docs を確認し、必要なら更新する
- DB 変更時は schema / init SQL / migration と DB 関連 docs を確認し、必要なら更新する
- 手順変更時は `README.md` または `docs/runbook.md` を更新する
- フェーズ進行、完了状態、次にやることが変わった場合は `docs/current-status.md` を更新する
- 優先順位が変わった場合は `docs/backlog-priority.md` を更新する
- 新しい制約や未解決事項が増えた場合は `docs/known-issues.md` を更新する
- 重要な利用者向け変更や運用上の変更は `CHANGELOG.md` を更新する
- エージェント向けの入口説明や恒久ルールが変わった場合は `AGENT.md` や `RULES.md` も更新する
- ドキュメント更新が不要と判断した場合でも、完了前に更新不要の理由を確認する

## Where To Start Next

次に着手する優先順は以下です。

1. kind / Helm / Argo CD の整備に着手する
2. Proxmox クラスタ移行、永続化、監視拡張を進める
3. 認証導入の前提整理を進める
4. collector の収集戦略拡張を進める
