# AGENT.md

## Responsibility

`article-service` は記事管理の中核サービスです。

- 記事一覧、詳細、検索を提供する
- collector からの内部取り込みを受け付ける
- ソース情報と取得ジョブ履歴の基礎データを管理する

## Read First

- `cmd/article-service/main.go`: 起動エントリポイント
- `internal/app/app.go`: DI と初期化
- `internal/httpapi/router.go`: HTTP ルート
- `internal/service/article_service.go`: ユースケース
- `internal/repository/`: 永続化
- `internal/domain/models.go`: ドメインモデル
- `internal/events/publisher.go`: イベント発行

## Implementation Notes

- API 変更は `router.go` と `article_service.go` を対で確認する
- データ更新を伴う変更は repository と domain の整合を崩さない
- 収集連携に影響する変更は collector 側の投入仕様も確認する
