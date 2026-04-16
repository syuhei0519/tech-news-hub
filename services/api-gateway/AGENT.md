# AGENT.md

## Read When

- 公開 API を追加、変更、削除するとき
- proxy、CORS、共通エラー、HTTP ミドルウェアを触るとき

## Open First

- `internal/httpapi/routes.go`
- `internal/httpapi/middleware.go`
- `internal/httpapi/routes_test.go`
- `internal/app/app.go`
- `../../docs/openapi/api-gateway.summary.md`

## Usually Skip

- frontend の表示実装
- collector / notification の内部詳細

ただし upstream 契約を変える場合は対象サービスの `AGENT.md` も開きます。
schema や parameter まで変える場合は `../../docs/openapi/api-gateway.yaml` も開きます。

## Minimum Verification

- `cd services/api-gateway && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...`
