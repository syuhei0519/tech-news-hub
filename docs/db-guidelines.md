# DB Guidelines

## Purpose

このドキュメントは、MySQL を前提としたデータ設計と運用方針を定義する。

## Basic Policy

- DB は MySQL を採用する
- 初期は 1 インスタンスで開始する
- ただしサービス責務はコード上で分離する
- 将来的な分離を見据えて、テーブル責務を明確に保つ

## Ownership Rules

- article-service が記事、ソース、取得ジョブ履歴の主要テーブルを扱う
- collector-service は DB を直接操作しないか、極小に留める
- notification-service は将来専用テーブルや配信履歴を持てるよう分離を意識する
- 共有 DB でも、他サービスの責務テーブルを横断的に触りすぎない

## Table Design Rules

- 主キーは `id BIGINT` を基本とする
- 監査用に `created_at`, `updated_at` を持つ
- 状態変化を持つテーブルは時刻と status を残す
- enum は将来変更しやすさを考え、文字列カラムでもよい
- nullable は意味が明確なものだけにする

## Article Table Rules

- URL と `dedupe_key` を保持する
- `dedupe_key` には unique 制約を付ける
- `source_id`, `published_at`, `category` には検索を見て index を検討する
- `is_read`, `is_favorite` は初期は記事テーブルに持ってよい
- タグは初期は JSON 配列で許容する

## Source Table Rules

- source 名は一意制約を検討する
- `is_enabled` を持たせる
- `last_fetched_at`, `last_fetch_status`, `last_error_message` を保持する
- `default_category`, `fetch_method`, `interval_minutes` を管理する

## Fetch Job Rules

- 取得開始時刻、終了時刻、status を必ず記録する
- `fetched_count`, `inserted_count`, `duplicated_count` を持つ
- エラーは要約メッセージを残す
- collector の再試行分析ができる粒度を残す

## Index Rules

- 一覧画面で使う列に優先して index を付ける
- まずは以下を優先候補とする

`articles.source_id`
`articles.published_at`
`articles.category`
`articles.dedupe_key`
`fetch_jobs.source_id`
`sources.name`

- 過剰に index を貼りすぎない
- クエリ実績を見て追加する

## Search Rules

- 初期は `title`, `excerpt` の検索を優先する
- MySQL の `LIKE` から始めてもよい
- 必要なら FULLTEXT INDEX を使う
- 検索の重さが出たら別手段を検討する

## Migration Rules

- 初期は Compose 用初期 SQL でもよい
- ただし中長期では migration 管理方式を導入する
- schema 変更時は rollback 可能性を考える
- k8s 運用を見据えて migration 実行方法を早めに決める

## Access Rules

- 学習効率優先なら GORM を使ってよい
- 実務寄りに寄せるなら sqlc + 生 SQL も有力
- このリポジトリではクエリ可視性を重視する
- どちらを使っても repository 層で責務を隠蔽する

## Consistency Rules

- 重複防止は DB 制約でも担保する
- アプリケーション判定だけに依存しない
- collector の再試行時に同一記事が増えないことを優先する

## Time Rules

- DB 保存時刻は UTC ベースにする
- 表示側でローカルタイムへ変換する
- `published_at` が欠けるケースを許容する
- `fetched_at` は必須とする

## CSV Rules

- CSV 向けクエリは一覧 API と条件整合を保つ
- 出力項目と並び順は固定する
- 大量データ時は件数上限または分割戦略を取る

## Future Separation Rules

- 将来的にサービスごとの DB 分離ができるようにする
- そのためにテーブル境界、repository 境界、internal API 境界を保つ
- 他サービスの都合で article-service のスキーマを安易に変更しない

## Backup / Persistence Rules

- ローカルでは単純構成でよい
- 実クラスタでは PVC とバックアップ方針を決める
- MySQL データの永続化は early stage から意識する

## Performance Rules

- ミニ PC 32GB 前提で無理のない構成にする
- 最初から大規模最適化しない
- 一覧と検索の体感速度を優先する
