# API Guidelines

## Purpose

このドキュメントは、`Tech Feed Hub` における API 設計の基本方針を定義する。

対象:

- frontend 向け公開 API
- サービス間の内部 API
- 将来の CSV / 認証 / ダッシュボード API

## General Rules

- 公開 API は `api-gateway` 経由で提供する
- 公開 API のパスは `/api/v1/...` に統一する
- サービス間通信用 API は `/internal/...` に分離する
- OpenAPI を先に意識して設計する
- レスポンス形式はできるだけ一貫させる
- 破壊的変更は version を上げるか、段階的に移行する

## API Boundary Rules

### frontend -> api-gateway

- frontend は原則として `api-gateway` のみを呼ぶ
- frontend から各マイクロサービスへ直接アクセスしない

### api-gateway -> internal services

- gateway は認証、ヘッダ、レスポンス整形、CSV レスポンス制御を担う
- gateway は業務ロジックの本体を持ちすぎない

### service-to-service

- 内部 API は最小限に保つ
- 依存方向は単純化する
- collector-service から article-service への登録依頼は許容する
- article-service から collector-service への依存は避ける

## Resource Design

- 複数形リソースを基本とする
- 一覧は `GET /api/v1/articles`
- 単体取得は `GET /api/v1/articles/{id}`
- 更新系は対象を明示する
- 動作系エンドポイントは名詞ベースを優先し、必要時のみ action を使う

例:

- `GET /api/v1/articles`
- `GET /api/v1/articles/{id}`
- `PATCH /api/v1/articles/{id}/read-status`
- `PATCH /api/v1/articles/{id}/favorite-status`
- `GET /api/v1/sources`
- `POST /api/v1/sources`
- `PATCH /api/v1/sources/{id}`
- `GET /api/v1/exports/articles.csv`

## Query Parameter Rules

- 検索、フィルタ、ソート、ページネーションは query parameter で表現する
- パラメータ名は省略しすぎない
- 真偽値は `true` / `false`
- 日付範囲は `from`, `to` など対になる名称を使う

例:

- `q`
- `category`
- `source_id`
- `tag`
- `is_read`
- `is_favorite`
- `sort`
- `order`
- `page`
- `page_size`
- `from`
- `to`

## Pagination Rules

- 一覧 API はページネーション前提で設計する
- 初期の page size は 20 を基本にする
- 上限を設ける
- レスポンスには `total`, `page`, `page_size`, `total_pages` を含める

## Sorting Rules

- `sort` と `order` を分ける
- `order` は `asc` / `desc`
- デフォルトソートは新着順
- ソート可能項目は OpenAPI に明示する

## Filtering Rules

- カテゴリ、ソース、タグ、未読、お気に入りは独立して指定可能にする
- 複合条件を前提に設計する
- 未指定時はフィルタしない

## Search Rules

- 初期はタイトルと抜粋を対象にする
- 検索対象拡張時も API パラメータ名はできるだけ維持する
- 文字列検索の実装方式は DB インデックスを見ながら調整する

## Response Rules

### Success

- `200 OK`: 取得成功
- `201 Created`: 作成成功
- `202 Accepted`: 非同期受付
- `204 No Content`: 本文不要の更新成功

### Error

- `400 Bad Request`: 入力不正
- `401 Unauthorized`: 認証失敗
- `403 Forbidden`: 認可拒否
- `404 Not Found`: リソース未存在
- `409 Conflict`: 重複や競合
- `422 Unprocessable Entity`: バリデーション不正
- `500 Internal Server Error`: 想定外障害
- `502 Bad Gateway`: upstream service failure

### Error Body

最低限、以下を含める。

```json
{
  "error": "validation_error",
  "message": "page_size must be between 1 and 100",
  "trace_id": "..."
}
```

## Validation Rules

- gateway と各 service の両方で必要なバリデーションを行う
- 必須、型、範囲、文字数、enum を明示する
- CSV 出力など重い API は件数上限を設ける

## Authentication Rules

- 公開 API は将来的に JWT 前提で設計する
- 認証不要な health check は分離する
- 初期の単一ユーザー前提でも、認証導入位置は gateway に固定する

## CSV Rules

- CSV 出力 API は gateway で公開する
- 実データ取得は article-service に委譲する
- UTF-8 を基本とする
- Excel 利用を意識したエスケープを行う
- 件数上限を設ける
- 並び順と日時フォーマットを固定する

## Internal API Rules

- internal API は外部公開用途に流用しない
- internal API では source metadata や job status 更新をまとめて扱ってよい
- internal API でも idempotency を意識する

## Idempotency Rules

- collector からの取り込みは重複登録を避ける
- `dedupe_key` を使って冪等性を担保する
- 再試行時に同一記事が多重登録されないことを優先する

## Observability Rules

- すべてのリクエストに trace_id を持たせやすい構造にする
- gateway はアクセスログを出す
- エラー時は upstream と request context が追える情報を残す

## OpenAPI Rules

- 新しい公開 API を追加する際は OpenAPI を更新する
- request / response schema を定義する
- query parameter と enum 値を明示する
- examples を付けられるものは付ける

## Naming Rules

- JSON key は snake_case を基本とする
- DB 項目名と API 項目名は極端に乖離させない
- ただし外部公開で不自然な名前は避ける
