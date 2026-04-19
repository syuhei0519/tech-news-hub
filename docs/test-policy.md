# Test Policy

## Purpose

このドキュメントは、どこにどの種類のテストを書くかの基準を定義する。

## Testing Principles

- 重要な責務境界を優先して検証する
- article-service の安定性を優先する
- 過剰な E2E より、保守しやすい単位で積む
- ローカルと CI で回しやすい構成を優先する

## Test Layers

### Unit Test

対象:

- service ロジック
- utility
- 正規化処理
- フィルタ条件組み立て
- dedupe 処理

優先度:

- 高い

### Component Test

対象:

- frontend の主要画面
- API client の query / payload 組み立て
- mutation 後の画面同期

優先度:

- 高い

### Integration Test

対象:

- repository
- DB アクセス
- API handler
- collector から article-service への連携

優先度:

- 高い

### End-to-End Test

対象:

- frontend と gateway の主要ユーザーフロー
- 記事一覧、詳細、検索、CSV
- 通知一覧、既読更新

優先度:

- 中

### Contract Test

対象:

- api-gateway ↔ article-service / notification-service
- collector-service ↔ article-service internal API
- RabbitMQ event payload ↔ notification-service

優先度:

- 高い

## Minimum Expectations By Area

### article-service

- 一覧取得
- 詳細取得
- 検索 / フィルタ / ソート
- 重複防止
- ソース状態更新
- fetch job 記録

### collector-service

- RSS 正規化
- dedupe_key 生成
- ingest payload 生成
- 失敗時のエラー扱い

### api-gateway

- upstream プロキシ
- エラー透過または整形
- query parameter 引き継ぎ

### frontend

- 主要画面のレンダリング
- 検索フォーム
- API レスポンスに対する表示
- 失敗表示
- 記事詳細の既読 / お気に入り更新
- 通知一覧の既読更新
- source 管理フォームと source 詳細

## When Tests Are Required

以下ではテスト追加を強く推奨する。

- 既存バグ修正
- DB クエリ変更
- フィルタ / ソート / 検索条件変更
- CSV 出力ロジック変更
- collector の正規化ルール変更
- 通知イベント契約変更
- frontend の検索条件、mutation、画面導線変更
- service 間 request / response 仕様変更

## When A Test May Be Deferred

以下は一時的に defer できる。

- 単純な文言修正
- docs のみの変更
- 将来削除予定の仮実装

ただし defer した場合は PR に明記する。

## CI Expectations

最低限 CI で回す対象:

- backend test
- frontend unit/component test
- frontend build
- compose config validation
- article-service MySQL integration test
- notification-service MySQL integration test

PR での扱い:

- `verify` は高速レーンとして常時実行し、`make verify` の責務を維持する
- `integration` は article-service と notification-service の MySQL integration test を常時実行する
- `smoke` は E2E レーンとして分離し、PR では frontend / gateway / article-service / notification-service / compose / workflow 変更時だけ実行する
- `main` への push と `workflow_dispatch` では、docs-only を除き `smoke` も実行する

補足:

- 通常のコード変更では `verify` に `make verify`、`integration` に `make test-article-integration` と `make test-notification-integration` を割り当てる
- docs-only 変更では、required status check を維持するため `verify` ジョブは成功させる
- ただし docs-only 変更時は重い `make verify` を省略してよい
- docs-only 判定条件は workflow 側で管理する
- `make verify` に E2E は含めず、`make e2e-up` / `make e2e-seed` / `make test-e2e` / `make e2e-down` で責務を分ける
- `integration` は MySQL service container 付きの独立 job とし、`verify` に混ぜない

将来的に追加候補:

- API / event contract test
- 他サービスの MySQL integration test 追加
- kind smoke test

## Minimal E2E Scope

Phase 4.5 で持つ E2E は次の 2 本に限定する。

1. 記事一覧表示 → 検索条件変更 → 詳細遷移 → CSV download
2. 通知一覧表示 → 既読更新

方針:

- 実行方式は compose ベースとする
- collector や外部 RSS には依存せず、MySQL に deterministic な seed を投入して再現する
- flaky 要因になりやすい RabbitMQ 到着待ち、外部 HTTP、現在時刻依存は E2E 導線に含めない

## Phase 4.5 Priority

Phase 5 より前に、次を優先して整備する。

1. article-service の repository / handler / MySQL integration test
2. frontend の記事一覧 / 詳細導線の component test
3. notification / source 管理周辺の test
4. collector のデータ揺れケースと service 間 contract test
5. 最小 E2E と CI への組み込み

## Test Data Policy

- 外部ソースへの依存を避ける
- collector テストでは固定 RSS サンプルを使う
- 時刻依存は固定化または注入する
- CSV は列順と日付形式を検証する

## Bugfix Policy

- バグ修正時は、可能なら再発防止テストを追加する
- 追加できない場合は理由を PR に記載する
