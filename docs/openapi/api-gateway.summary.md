# API Gateway OpenAPI Summary

## Purpose

このファイルは、通常セッションで `docs/openapi/api-gateway.yaml` を毎回開かずに公開 API の全体像を把握するための要約です。

次のケースでは summary だけを読む:

- 既存 endpoint の所在を確認したい
- frontend / gateway の参照先を素早く確認したい
- API 変更の影響範囲をざっくり見たい

次のケースでは `docs/openapi/api-gateway.yaml` も開く:

- request / response schema を変える
- query parameter や status code を追加、変更、削除する
- OpenAPI 本体の更新漏れがないか確認する

## Endpoint Groups

### Health

- `GET /health`

### Articles

- `GET /api/v1/articles`
  - query: `q`, `category`, `source_id`, `is_read`, `is_favorite`, `from`, `to`, `page`, `page_size`
- `GET /api/v1/articles/{id}`
- `PATCH /api/v1/articles/{id}/read-status`
- `PATCH /api/v1/articles/{id}/favorite-status`
- `GET /api/v1/exports/articles.csv`
  - query: article list に加えて `sort`, `order`

### Sources

- `GET /api/v1/sources`
- `POST /api/v1/sources`
- `GET /api/v1/sources/{id}`
- `PATCH /api/v1/sources/{id}`
- `DELETE /api/v1/sources/{id}`
- `GET /api/v1/fetch-jobs`
  - query: `source_id`, `status`, `page`, `page_size`

### Notifications

- `GET /api/v1/notifications`
  - query: `is_read`, `page`, `page_size`
- `PATCH /api/v1/notifications/{id}/read-status`

## Notes

- frontend からの入口は `api-gateway` に集約する
- source 管理と notification API も gateway 経由で公開する
- 詳細 schema は `docs/openapi/api-gateway.yaml` を正本として扱う
