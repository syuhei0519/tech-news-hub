# AGENT.md

## Purpose

このファイルは、このリポジトリで作業を始めるときの最短ルータです。
最初から全ドキュメントを読まず、必要な範囲だけ開く前提で使います。

## Always Read

作業開始時は、まず次だけ確認します。

- `AGENT.md`
- `docs/current-status.md`
- 対象領域の `AGENT.md`

必要なら追加で確認するもの:

- `docs/agent-playbooks.md` の該当セクション
- `docs/openapi/api-gateway.summary.md`: 公開 API の全体像だけ先に見たいとき

対象領域:

- `frontend/AGENT.md`
- `services/api-gateway/AGENT.md`
- `services/article-service/AGENT.md`
- `services/collector-service/AGENT.md`
- `services/notification-service/AGENT.md`

## Read Only If Needed

- `RULES.md`: スコープ、責務、全体方針を変えるとき
- `docs/backlog-priority.md`: 次に着手する Issue を選ぶとき
- `docs/test-policy.md`: テスト追加判断に迷うとき
- `docs/runbook.md`: 起動、切り分け、運用手順が必要なとき
- `docs/branch-strategy.md`: branch、PR、merge の運用確認が必要なとき
- `docs/api-guidelines.md`: 公開 API を変えるとき
- `docs/db-guidelines.md`: DB や repository の責務を変えるとき
- `docs/k8s-guidelines.md`: Helm、Argo CD、Kubernetes を変えるとき
- `docs/requirements.md`: 要件やフェーズ適合を確認したいとき
- `docs/openapi/api-gateway.yaml`: schema や parameter まで変更確認したいとき

## Quick Router

- UI 修正、表示崩れ、検索条件変更: `frontend/AGENT.md`
- 公開 API 追加、gateway ルート変更: `services/api-gateway/AGENT.md`
- 記事、ソース、fetch job の振る舞い変更: `services/article-service/AGENT.md`
- 収集、正規化、ingest、dedupe 変更: `services/collector-service/AGENT.md`
- 通知 API、既読、イベント消費変更: `services/notification-service/AGENT.md`
- docs-only 変更: `docs/agent-playbooks.md` の `Docs-Only`
- endpoint 一覧や影響範囲だけ先に確認したい: `docs/openapi/api-gateway.summary.md`

## Non-Negotiables

- Git のコミットメッセージは日本語で記述する
- `feat:` `fix:` `docs:` などの種別プレフィックスは英語でもよいが、説明部分は日本語にする
- 変更に対応する docs は同じ変更内で更新する
- PR の題名と本文は、日本語指定がない限り日本語で書く
- PR 作成時は `.github/pull_request_template.md` を基準に本文を組み立てる
- GitHub API / MCP で PR を作る場合、テンプレートは自動適用されない前提で `.github/pull_request_template.md` を開いて手動で埋める
- 明示的な指示がない限り `main` へ直接 push しない
- 重要な方針変更は `RULES.md` 側も更新対象として確認する

## Verification Policy

- まず対象領域の最小検証を実行する
- 複数領域にまたがる変更、または最小検証で不安が残る変更では `make verify` を使う
- docs-only 変更では重い検証を省略してよい

## Documentation Router

- 現在地、完了状況、次優先が変わる: `docs/current-status.md`
- 優先順位が変わる: `docs/backlog-priority.md`
- API 契約が変わる: まず `docs/openapi/api-gateway.summary.md`、schema 更新時は `docs/openapi/api-gateway.yaml` と API 関連 docs
- DB / schema / migration 方針が変わる: DB 関連 docs
- 起動や運用手順が変わる: `README.md` または `docs/runbook.md`
- 恒久ルールが変わる: `RULES.md`
- エージェントの入口や最小読込導線が変わる: `AGENT.md` または `docs/agent-playbooks.md`
