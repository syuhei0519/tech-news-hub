# 図ドキュメント

このディレクトリには、ルート repo の現行実装を前提にした Mermaid 図をまとめています。
README や個別 docs から辿れるようにし、runtime 構成、データ構造、主要フロー、CI / テストレーンを短時間で把握できるようにするのが目的です。

今回の図はこのルート repo を対象にしており、`tech-news-hub/` 配下は含めていません。
また、kind / Helm / Argo CD、監視基盤、Redis など repo 上で未実装の将来要素は図に含めていません。

## 一覧

- [システム全体構成図](system-overview.md)
  - frontend、api-gateway、各 service、MySQL、RabbitMQ、外部 RSS の関係を示します。
- [データモデル / 所有責務 ER 図](data-model-ownership-er.md)
  - `sources`、`articles`、`fetch_jobs`、`notifications` の関係、FK / unique 制約、service ごとの所有責務を示します。
- [記事収集フロー図](article-collection-flow.md)
  - `collector-service` が source 同期、fetch job 管理、RSS 正規化、ingest をどう進めるかを示します。
- [通知生成フロー図](notification-generation-flow.md)
  - `article.ingested` と `collector.fetch.failed` が通知一覧に見えるまでの非同期経路を示します。
- [公開 API / 内部 API 責務境界図](api-boundary-public-internal.md)
  - frontend route、`api-gateway` の公開 API group、upstream service の `/api/v1` と `article-service` の `/internal` 境界を示します。
- [CI / テストレーン構成図](ci-test-lanes.md)
  - `changes -> verify / integration / coverage / smoke` の条件分岐と責務を示します。

## 補足

- 図には repo から断定できるものだけを載せています。
- 曖昧な点は図本体には描かず、各ファイルの注記に `repo からは未確定` として分けています。
