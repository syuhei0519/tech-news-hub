# Phase 4.5 改善計画

## 目的

`improvement-task-list-tech-news-hub.md` をベースに、現在の実装状況を踏まえて Phase 4.5 の改善を実行計画に落とし込む。

この計画の主眼は、新機能追加ではなく、既存価値を壊さないための品質ゲート、テスト責務、回帰防止基盤の強化である。

## 計画の前提

2026-04-19 時点の前提:

- `make verify` は backend test + frontend test + build + compose config check を実行する
- frontend の component test 基盤は導入済み
- article-service の repository / handler integration test はコード上に存在する
- Playwright E2E の最小 smoke と CI レーンは導入済み
- article-service と notification-service の integration test は CI 常設ジョブで実行される
- collector の契約テスト、notification-service の DB integration test、主要 API の error case test は実装済み
- frontend test helper と fixture 方針はコード上に導入済み
- `#28` の親 Issue は completed でクローズ済みであり、Phase 4.5 は完了扱いとする

したがって、この文書の主目的は未着手タスクの計画ではなく、Phase 4.5 で実施した内容を将来参照できる形で固定することにある。

## 到達目標

Phase 4.5 完了時に、次の状態を作る。

1. 日常の検証コマンドが明確で、ローカルと CI の責務が揃っている
2. article-service / collector-service / notification-service / frontend の主要責務境界が自動検証される
3. E2E は最小本数に絞りつつ、再現性高く回る
4. テスト追加時の作法とデータ戦略が文書化され、継続拡張しやすい

## 非目標

この計画では、次は優先しない。

- Phase 5 の kind / Helm / Argo CD の本格整備
- 認証導入
- 大規模な UI 改修や機能追加
- 監視、バックアップ、本番運用の詳細設計

## 実行方針

- 先に品質ゲートへ組み込めるものから着手し、テスト資産を常設化する
- article-service を最優先にしつつ、利用者から見える HTTP 契約と service 間契約を固定する
- E2E は薄く保ち、integration / contract / component に責務を寄せる
- docs は実装と同じ変更で更新し、運用ルールと乖離させない

## 実施フェーズ

各フェーズの主要子 Issue はすでにクローズ済みである。以下は「予定」ではなく、実施内容と残件整理として読む。

### フェーズ 1: 品質ゲート統合

#### 狙い

既にあるテストを日常導線に組み込み、「書いたが回っていない」状態をなくす。

#### 対象タスク

- `make test-frontend` を追加し、`frontend` の `npm run test:run` を実行可能にする
- `make verify` に frontend unit/component test を組み込む
- article-service integration test を `make test-article-integration` で常設運用しやすくする
- CI を `fast` / `integration` / `smoke` の 3 層に整理する
- README と `docs/test-policy.md` と `docs/runbook.md` のコマンド説明を同期する

#### 成果物

- 更新された `Makefile`
- 更新された `.github/workflows/ci.yml`
- 更新された README / test policy / runbook

#### 完了条件

- `make verify` で backend test + frontend test + build + compose config check が通る
- CI の job 名と責務が見て分かる
- `integration` job で article-service integration test が常時実行される

### フェーズ 2: 重要境界の契約固定

#### 狙い

主要な仕様破壊が DB 境界、HTTP 境界、service 間境界で検知できる状態を作る。

#### 対象タスク

- article-service handler integration test を一覧 / 詳細 / 更新 / CSV まで拡充する
- collector-service から article-service への ingest 契約テストを追加する
- notification-service の DB integration test を追加する
- API エラー系のテストを追加する

#### 優先順

1. article-service handler integration
2. collector -> article-service contract
3. notification-service DB integration
4. API error cases

#### 成果物

- article-service の HTTP integration test 拡充
- collector-service の contract test
- notification-service の repository or DB integration test
- エラー系テスト追加

#### 完了条件

- 主要 API の正常系 / 異常系が HTTP レベルで固定される
- collector の payload 変更が CI で検知できる
- notification-service の永続化境界が自動で守られる

### フェーズ 3: テスト保守性の改善

#### 狙い

テスト件数が増えても重複、fixture の肥大化、画面ごとの個別実装を抑える。

#### 対象タスク

- frontend の `renderWithProviders`、MSW handler、fixture builder を共通化する
- テストデータ戦略を文書化する
- E2E seed と fixture を最小構成へ整理する
- 時刻依存や外部依存をさらに固定する

#### 成果物

- frontend test helper 群
- テストデータ方針の文書
- E2E seed / fixture 整理

#### 完了条件

- 新規画面テスト追加時に共通 helper を再利用できる
- fixture の置き方と命名が統一される
- E2E の flaky が目立つ状態でなくなる

### フェーズ 4: 可視化と差分検知の強化

#### 狙い

どこが未保護か、どこで仕様差分が出たかを CI で見えるようにする。

#### 対象タスク

- Go / frontend coverage を artifact 化する
- OpenAPI lint と breaking change check の導入方針を固める
- 実装と docs の差分検知を自動化する

#### 成果物

- coverage 出力付き CI
- OpenAPI 差分検知ジョブ

#### 完了条件

- PR 単位で未テスト領域を把握しやすい
- API 変更時に docs 更新漏れを検知できる

## 実施順序

最短で効果が出る順序は次の通りとする。

1. `make verify` への frontend test 組み込み
2. CI 3 層化と article integration 常設
3. article-service handler integration 拡充
4. collector 契約テスト
5. notification-service DB integration
6. frontend test 共通基盤整理
7. E2E seed 安定化
8. エラー系テスト拡充
9. テストデータ方針の文書化
10. coverage 可視化
11. OpenAPI 差分検知

## Issue 対応関係

- `#28`: Phase 4.5 全体親 Issue。completed で closed
- `#40`: Milestone A の親 Issue。closed
- `#41`: Milestone B の親 Issue。closed
- `#42`: Milestone C の親 Issue。closed
- `#43`: フェーズ 1 の frontend test 導線化。closed
- `#44`: フェーズ 1 の CI 3 層化。closed
- `#45`: フェーズ 1 の article integration 常設。closed
- `#46`: フェーズ 2 の article handler integration。closed
- `#47`: フェーズ 2 の collector contract test。closed
- `#48`: フェーズ 2 の notification-service DB integration。closed
- `#49`: フェーズ 2 の API error cases。closed
- `#50`: フェーズ 3 の frontend test 共通基盤整理。closed
- `#51`: フェーズ 3 とフェーズ 4 の E2E 安定化 / 可視化。closed

## 推奨マイルストーン

### マイルストーン A

品質ゲートの再整理を終える。

- `make verify` に frontend test を統合
- CI 3 層化
- article integration 常設

### マイルストーン B

主要境界の自動検証を終える。

- article handler integration
- collector contract test
- notification DB integration
- 主要 API の error cases

### マイルストーン C

保守性と可視化を仕上げる。

- frontend test helper 整理
- E2E seed 安定化
- test data policy 文書化
- coverage / OpenAPI 差分検知

## リスクと対策

- CI 時間増加
  - `fast` と `integration` を分離し、E2E は `smoke` に限定する
- MySQL 依存テストの不安定化
  - migration、seed、時刻依存、起動待ちを共通化する
- テスト追加で fixture が肥大化する
  - factory / seed / static fixture の使い分けを先に文書化する
- ドキュメントが実装に追従しない
  - Makefile / workflow / test policy / runbook を同一 PR で更新する

## Phase 4.5 の完了判定

次を満たしたら、Phase 4.5 は完了とみなす。

- `make verify` が frontend test を含む標準検証導線として機能している
- CI が `fast` / `integration` / `smoke` の責務で整理されている
- article-service、collector-service、notification-service の主要境界に自動テストがある
- E2E smoke が再現性高く維持されている
- テストポリシーとテストデータ方針が文書化されている
- coverage または OpenAPI 差分検知の少なくとも一方が CI に入っている

補足:

- 子 Issue `#40` から `#51` と親 Issue `#28` はクローズ済みであり、Phase 4.5 は完了したものとして扱う
- 次の優先フェーズは `#7` の kind / Helm / Argo CD 整備である
