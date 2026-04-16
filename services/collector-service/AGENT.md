# AGENT.md

## Responsibility

`collector-service` は外部ソース収集と正規化を担当します。

- RSS などの外部データを取得する
- 記事データを正規化し `dedupe_key` を生成する
- article-service へ取り込みを依頼する
- 必要に応じてイベントを発行する

## Read First

- `cmd/collector-service/main.go`: 起動エントリポイント
- `internal/app/app.go`: 設定と初期化
- `internal/httpapi/routes.go`: 実行トリガーの HTTP 入口
- `internal/collector/service.go`: 収集と正規化の本体
- `internal/events/publisher.go`: イベント発行
- `internal/collector/service_test.go`: 収集ロジックの期待仕様
- `internal/collector/service_http_test.go`: HTTP 経由の期待仕様

## Implementation Notes

- 収集ロジック変更時は dedupe の安定性を崩さない
- article-service への投入契約を変える場合は両サービスを同時に確認する
- 外部入力の揺れは `internal/collector/service.go` で吸収する
