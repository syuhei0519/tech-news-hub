# Rules Operations

## Purpose

このドキュメントは、このリポジトリで定義したルール群をどう運用するかをまとめたものです。

対象:

- `RULES.md`
- `AGENT.md`
- `docs/requirements.md`
- `docs/api-guidelines.md`
- `docs/db-guidelines.md`
- `docs/k8s-guidelines.md`
- GitHub Issue / PR テンプレート

## Source Of Truth

ルール参照の優先順位は以下です。

1. `RULES.md`
2. `docs/requirements.md`
3. `docs/api-guidelines.md`
4. `docs/db-guidelines.md`
5. `docs/k8s-guidelines.md`
6. `docs/architecture.md`
7. `docs/phases.md`
8. `AGENT.md`

補足:

- `RULES.md` は最上位の運用ルール
- `docs/requirements.md` は要件の基準
- 各 guideline は領域ごとの実装判断基準
- `AGENT.md` は開発者向けの入口

## How To Use

### When Starting Work

作業開始時は最低限、以下を確認する。

- `AGENT.md`
- `RULES.md`
- 今回の変更に関係する guideline

例:

- API 変更なら `docs/api-guidelines.md`
- DB 変更なら `docs/db-guidelines.md`
- Kubernetes / Helm 変更なら `docs/k8s-guidelines.md`

### When Designing

- まず要件が `docs/requirements.md` に反していないか確認する
- 責務境界が `RULES.md` の方針を壊していないか確認する
- フェーズ外の機能を入れようとしていないか確認する

### When Implementing

- API を追加したら OpenAPI 更新対象か確認する
- DB スキーマを変えたら migration / init SQL / ドキュメントの整合を確認する
- Kubernetes 関連を変えたら Helm values で環境差分吸収できるか確認する

### When Reviewing

レビュー時は以下を確認する。

- ルール違反がないか
- 責務分離が崩れていないか
- 初期スコープ外の複雑化が入っていないか
- ドキュメント更新漏れがないか

## Update Policy

### RULES.md を更新する場合

以下のような変更時に更新する。

- プロジェクト全体方針の変更
- 初期スコープ / 非スコープの変更
- サービス責務の変更
- フェーズ順序の変更
- 運用方針の変更

### requirements.md を更新する場合

以下のような変更時に更新する。

- 要件追加 / 削除
- 対象ソースの追加
- 非機能要件や学習要件の変更
- 技術スタック方針の変更

### api-guidelines.md を更新する場合

以下のような変更時に更新する。

- API 命名規則の変更
- エラーレスポンス方針の変更
- CSV API 方針の変更
- 認証 / versioning 方針の変更

### db-guidelines.md を更新する場合

以下のような変更時に更新する。

- テーブル責務変更
- migration 方針変更
- ORM / sqlc 方針変更
- index / 検索方針変更

### k8s-guidelines.md を更新する場合

以下のような変更時に更新する。

- Helm / Argo CD 方針変更
- Ingress / Secret / PVC 方針変更
- kind から実クラスタ移行戦略変更
- リソース配分方針変更

## Change Checklist

変更内容ごとに、最低限以下を揃える。

### API 変更

- 実装
- OpenAPI
- `docs/api-guidelines.md` との整合確認
- frontend 影響確認

### DB 変更

- スキーマ変更
- 初期 SQL または migration 更新
- repository / service 更新
- `docs/db-guidelines.md` との整合確認

### Kubernetes / Helm 変更

- manifest / chart / values 更新
- kind と実クラスタ差分の確認
- `docs/k8s-guidelines.md` との整合確認

### プロジェクト方針変更

- `RULES.md` 更新
- `docs/requirements.md` 更新
- 必要なら `AGENT.md` 更新

## GitHub Workflow Usage

### Issue

- バグは `bug_report.md`
- 機能追加は `feature_request.md`
- 実装タスクは `task.md`

Issue 作成時は、ルールや要件に関係する背景を記載する。

### Pull Request

- PR では `.github/pull_request_template.md` を使う
- `Scope`
- `Verification`
- `API / DB / Infra Impact`
- `Documentation`
  を最低限埋める

## Decision Rules

判断に迷った場合は次の順で決める。

1. `RULES.md` に反していないか
2. 初期スコープに収まっているか
3. Docker Compose で扱いやすいか
4. kind / Helm / Argo CD へ持っていきやすいか
5. ミニ PC 32GB で無理がないか
6. 学習効果が高いか

## Anti-Patterns

- ルールを更新せずに設計だけ変えること
- guideline を読まずに領域横断変更を進めること
- 初期スコープ外の複雑機能を黙って入れること
- Compose 前提を壊す変更を無言で入れること
- kind 専用実装を本番設計に埋め込むこと

## Practical Workflow

推奨運用手順:

1. Issue を作る
2. 影響範囲を確認する
3. 対応する rule / guideline を読む
4. 実装する
5. 必要な docs を更新する
6. 検証する
7. PR を作る

## Notes

- ルールは実装を止めるためではなく、判断を揃えるために使う
- ルールが現実に合わなくなったら、実装側で黙って逸脱せず、ルール側を更新する
