# AGENT.md

## Responsibility

`api-gateway` はフロント向けの単一入口です。

- 公開 API を受け、バックエンドサービスへ中継する
- 共通のログ、エラー応答、CORS など HTTP 境界の責務を持つ
- 将来の認証導入ポイントになる

## Read First

- `cmd/api-gateway/main.go`: 起動エントリポイント
- `internal/app/app.go`: アプリ初期化
- `internal/httpapi/routes.go`: ルーティング
- `internal/httpapi/middleware.go`: 共通ミドルウェア
- `internal/httpapi/routes_test.go`: ルートの期待仕様

## Implementation Notes

- 公開 API の追加や変更はまず `routes.go` を確認する
- 横断的な HTTP 振る舞いは `middleware.go` に寄せる
- サービス間 API を直接ここへ混在させず、公開境界として扱う
