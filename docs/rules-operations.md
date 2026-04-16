# Rules Operations

## Purpose

このドキュメントは、ルール群をどの順で、どの粒度で使うかを固定するためのものです。
トークン効率のため、常に全件読まず、必要な文書だけ開きます。

## Managed Documents

- `RULES.md`
- `AGENT.md`
- `docs/agent-playbooks.md`
- `docs/requirements.md`
- `docs/api-guidelines.md`
- `docs/db-guidelines.md`
- `docs/k8s-guidelines.md`
- `docs/dod.md`
- `docs/runbook.md`
- `docs/test-policy.md`
- `docs/naming-conventions.md`
- `docs/release-migration-policy.md`
- `docs/branch-strategy.md`
- `docs/adr/*`
- `docs/current-status.md`
- `docs/backlog-priority.md`
- `docs/known-issues.md`
- `docs/glossary.md`
- `CHANGELOG.md`
- GitHub Issue / PR テンプレート

## Source Of Truth

優先順位は次の通りです。

1. `RULES.md`
2. `docs/requirements.md`
3. 領域別 guideline
4. `docs/dod.md`
5. `docs/test-policy.md`
6. `docs/release-migration-policy.md`
7. `docs/architecture.md`
8. `docs/phases.md`
9. `AGENT.md`
10. `docs/agent-playbooks.md`
11. `docs/current-status.md`

補足:

- `RULES.md` は全体方針
- `AGENT.md` は入口
- `docs/agent-playbooks.md` は最小読込の実務手順
- `docs/current-status.md` は次セッションの現在地

## Start Rules

### Always Read

作業開始時は次だけ確認する。

- `AGENT.md`
- `docs/current-status.md`
- 対象領域の `AGENT.md`
- `docs/agent-playbooks.md` の該当セクション

### Read On Demand

以下は必要になった時だけ開く。

- `RULES.md`: スコープ、責務、全体方針を変える時
- `docs/backlog-priority.md`: 次の Issue を選ぶ時
- `docs/branch-strategy.md`: branch / PR / merge 運用を確認する時
- `docs/test-policy.md`: テスト追加判断に迷う時
- `docs/runbook.md`: 起動、障害切り分け、運用手順が必要な時
- `docs/api-guidelines.md`: 公開 API を変える時
- `docs/db-guidelines.md`: DB、repository、migration を変える時
- `docs/k8s-guidelines.md`: Kubernetes、Helm、Argo CD を変える時
- `docs/requirements.md`: 要件適合に不安がある時

## How To Use

### When Designing

- 変更が `RULES.md` と `docs/requirements.md` に反していないか確認する
- プレイブックで十分に読める範囲を超える時だけ追加 docs を開く
- 重要な設計判断は ADR 化対象か確認する

### When Implementing

- API 変更時は OpenAPI 更新対象か確認する
- DB 変更時は init SQL / migration / repository 整合を確認する
- docs 更新対象を同じ変更内で処理する
- 処理意図が追いにくい箇所には日本語コメントを補う

### When Reviewing

- ルール違反がないか
- プレイブックに反して不要な横断変更が入っていないか
- ドキュメント更新漏れがないか
- `docs/dod.md` を満たしているか

## Change Checklist

### API Change

- 実装
- OpenAPI
- `docs/api-guidelines.md` との整合確認
- frontend 影響確認

### DB Change

- スキーマ変更
- 初期 SQL または migration 更新
- repository / service 更新
- `docs/db-guidelines.md` との整合確認

### Kubernetes / Helm Change

- manifest / chart / values 更新
- kind と実クラスタ差分の確認
- `docs/k8s-guidelines.md` との整合確認

### Project Policy Change

- `RULES.md` 更新
- 必要に応じて `docs/requirements.md` 更新
- 必要に応じて `AGENT.md` 更新
- エージェントの読込導線が変わるなら `docs/agent-playbooks.md` 更新
- 必要に応じて ADR または `CHANGELOG.md` 更新

## Update Policy

### Update `RULES.md` When

- プロジェクト全体方針が変わる
- 初期スコープや非スコープが変わる
- サービス責務やフェーズ順が変わる

### Update `AGENT.md` When

- 最初に読むべき導線が変わる
- 非交渉ルールを入口に残す必要が変わる

### Update `docs/agent-playbooks.md` When

- よくある作業の最小読込セットが変わる
- 新しい定番作業を追加したい
- 無駄な横断読込を減らしたい

### Update `docs/current-status.md` When

- 主要機能が完了した
- 現在フェーズが変わった
- 次優先が変わった

### Update `docs/backlog-priority.md` When

- 優先順位を入れ替えた
- 新しい高優先タスクを追加した

### Update `docs/known-issues.md` When

- 既知の制約や不具合を見つけた
- 既知課題が解消された

## GitHub Workflow Usage

### Issue

- バグは `bug_report.md`
- 機能追加は `feature_request.md`
- 実装タスクは `task.md`

### Pull Request

- `.github/pull_request_template.md` を使う
- `Scope`
- `Verification`
- `API / DB / Infra Impact`
- `Documentation`

## Decision Rules

迷った時は次の順で判断する。

1. `RULES.md` に反していないか
2. 初期スコープに収まっているか
3. Compose で扱いやすいか
4. kind / Helm / Argo CD へ持っていきやすいか
5. 学習効果が高いか

## Anti-Patterns

- 毎回フルドキュメント読込から始めること
- 領域外のコードや docs を無差別に開くこと
- ルール変更を docs に反映しないこと
- 重要な判断理由を ADR に残さないこと
- 現在地が変わったのに `docs/current-status.md` を更新しないこと

## Session Continuity

- セッション開始時は `docs/current-status.md` を確認する
- 作業ごとに `docs/agent-playbooks.md` の該当セクションを使う
- 優先順位判断は `docs/backlog-priority.md` を使う
- 既知制約は `docs/known-issues.md` を確認する
- 標準検証は `make verify` を使う
- Git 運用は `docs/branch-strategy.md` を参照する
