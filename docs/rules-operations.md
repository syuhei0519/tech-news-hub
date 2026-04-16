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

ルール参照の優先順位は以下です。

1. `RULES.md`
2. `docs/requirements.md`
3. `docs/api-guidelines.md`
4. `docs/db-guidelines.md`
5. `docs/k8s-guidelines.md`
6. `docs/dod.md`
7. `docs/test-policy.md`
8. `docs/release-migration-policy.md`
9. `docs/architecture.md`
10. `docs/phases.md`
11. `AGENT.md`

補足:

- `RULES.md` は最上位の運用ルール
- `docs/requirements.md` は要件の基準
- 各 guideline は領域ごとの実装判断基準
- `docs/dod.md` は完了判断基準
- `docs/runbook.md` は日常運用と切り分け手順
- `docs/test-policy.md` はテスト追加判断基準
- `docs/naming-conventions.md` は命名の基準
- `docs/release-migration-policy.md` は変更の安全な進め方
- `docs/branch-strategy.md` は Git 運用の基準
- `docs/adr/*` は設計判断の理由
- `docs/current-status.md` は次セッションの入口
- `docs/backlog-priority.md` は優先順の固定
- `docs/known-issues.md` は既知制約の固定
- `docs/glossary.md` は用語の固定
- `CHANGELOG.md` は重要変更の要約
- `AGENT.md` は開発者向けの入口

## How To Use

### When Starting Work

作業開始時は最低限、以下を確認する。

- `AGENT.md`
- `RULES.md`
- `docs/current-status.md`
- `docs/backlog-priority.md`
- `docs/branch-strategy.md`
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
- 重要な設計判断は ADR 化対象か確認する
- 既知制約なら `docs/known-issues.md` 更新対象か確認する
- 重要変更なら `CHANGELOG.md` 更新対象か確認する
- 処理意図が読み取りにくい実装には、明示指示がなくても日本語コメントを補う
- コメントは「なぜそうしているか」を優先し、逐語説明や一時メモは残さない

### When Reviewing

レビュー時は以下を確認する。

- ルール違反がないか
- 責務分離が崩れていないか
- 初期スコープ外の複雑化が入っていないか
- ドキュメント更新漏れがないか
- `docs/dod.md` を満たしているか
- PR の題名と本文が、明示的な指示なしで英語になっていないか
- 変更箇所のうち意図が追いにくい部分に、日本語コメントが不足していないか

### CI Interpretation

- 通常のコード変更では `make verify` を回す
- docs-only 変更では、required status check を壊さない範囲で重い検証を省略してよい
- その場合でも `verify` チェック自体は成功状態を返す
- docs-only 判定ロジックは `.github/workflows/ci.yml` を正本とする

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

### dod.md を更新する場合

以下のような変更時に更新する。

- 完了条件の見直し
- PR で毎回確認すべき事項の変更

### runbook.md を更新する場合

以下のような変更時に更新する。

- 起動手順の変更
- 障害切り分け手順の変更
- 日常運用手順の追加

### test-policy.md を更新する場合

以下のような変更時に更新する。

- テスト粒度や優先順位の変更
- CI 対象の変更

### naming-conventions.md を更新する場合

以下のような変更時に更新する。

- API / DB / code の命名方針変更
- branch / commit 命名方針変更

### release-migration-policy.md を更新する場合

以下のような変更時に更新する。

- 破壊的変更の扱い変更
- リリース / 移行の進め方変更

### branch-strategy.md を更新する場合

以下のような変更時に更新する。

- main 運用ルール変更
- PR 必須ルール変更
- merge 方式変更
- branch 種別変更
- branch 削除ルール変更

### ADR を追加する場合

以下のような変更時に追加する。

- 重要な設計採用判断
- 将来の実装方針へ長く影響する決定
- 単純な実装差分ではなく、判断理由を残す価値が高い変更

### current-status.md を更新する場合

以下のような変更時に更新する。

- 主要機能が完了した
- 現在フェーズが変わった
- 次にやるべき優先事項が変わった

### backlog-priority.md を更新する場合

以下のような変更時に更新する。

- 優先順位を入れ替えた
- 新しい高優先度 Issue を追加した

### known-issues.md を更新する場合

以下のような変更時に更新する。

- 既知の制約や不具合を見つけた
- 既知課題が解消された

### glossary.md を更新する場合

以下のような変更時に更新する。

- 用語の意味がぶれそうな概念を追加した
- 新しい内部概念や運用用語が増えた

### CHANGELOG.md を更新する場合

以下のような変更時に更新する。

- 重要な機能追加
- 重要な方針変更
- 次セッションで把握すべき大きな変更

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
- 必要なら ADR 追加
- 必要なら `CHANGELOG.md` 更新

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
- 重要な判断理由を ADR に残さないこと
- 現在地が変わったのに `docs/current-status.md` を放置すること
- 既知課題を Issue や口頭だけに残して `docs/known-issues.md` に書かないこと

## Practical Workflow

推奨運用手順:

1. Issue を作る
2. branch を切る
3. 影響範囲を確認する
4. 対応する rule / guideline を読む
5. 実装する
6. 必要な docs を更新する
7. `make verify` を実行する
8. PR を作る

## Session Continuity Rules

新規セッションでも品質を維持するために、次を運用する。

- セッション開始時に `docs/current-status.md` を確認する
- 優先順位判断は `docs/backlog-priority.md` を使う
- 既知制約は `docs/known-issues.md` を確認する
- 用語の意味は `docs/glossary.md` を確認する
- 大きな変更は `CHANGELOG.md` に残す
- 標準検証は `make verify` を使う
- コミットメッセージ規約は `docs/naming-conventions.md` を参照し、日本語で記述する
- Git 運用は `docs/branch-strategy.md` を参照する
- docs-only 変更時の CI 軽量化ルールは `.github/workflows/ci.yml` と `docs/test-policy.md` を参照する

## Notes

- ルールは実装を止めるためではなく、判断を揃えるために使う
- ルールが現実に合わなくなったら、実装側で黙って逸脱せず、ルール側を更新する
