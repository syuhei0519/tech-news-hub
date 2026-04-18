# Naming Conventions

## Purpose

このドキュメントは、リポジトリ内の命名規則を揃えるための基準を定義する。

## General Rules

- 意味が分かる名前を使う
- 省略しすぎない
- 同じ概念には同じ名前を使う
- API、DB、コード間で極端に名称をずらさない

## API Naming

- パスは複数形リソースを基本にする
- 公開 API は `/api/v1/...`
- 内部 API は `/internal/...`
- JSON key は `snake_case`

例:

- `source_id`
- `published_at`
- `is_favorite`
- `last_error_message`

## DB Naming

- テーブル名は複数形
- カラム名は `snake_case`
- 外部キーは `<resource>_id`
- 真偽値は `is_` プレフィックスを基本にする
- 時刻は `_at` サフィックスにする

例:

- `articles`
- `sources`
- `fetch_jobs`
- `is_enabled`
- `created_at`

## Go Naming

### Packages

- 小文字
- 短く、責務が分かる名前

例:

- `service`
- `repository`
- `domain`
- `httpapi`

### Files

- `snake_case.go`
- 役割が分かる単位で分ける

例:

- `article_service.go`
- `source_repository.go`
- `routes.go`

### Types

- 公開型は `PascalCase`
- private helper は必要以上に公開しない

例:

- `ArticleService`
- `ListArticlesParams`

### Methods / Functions

- 動詞から始める
- 一貫した語彙を使う

例:

- `ListArticles`
- `GetArticle`
- `EnsureSource`
- `UpdateFetchStatus`

## Frontend Naming

### Components

- `PascalCase.tsx`
- 画面は `*Page.tsx`
- レイアウトは `*Layout.tsx`

例:

- `ArticleListPage.tsx`
- `ArticleDetailPage.tsx`
- `AppLayout.tsx`

### Hooks / Store

- hook は `use` で始める
- Zustand store は `*Store.ts`

例:

- `useSearchStore`
- `searchStore.ts`

### Utility / API

- API client は `api.ts`
- ドメイン別に増える場合は役割単位で分割する

## Git Naming

### Branch

- 目的が分かる短い名前
- 詳細なブランチ運用は `docs/branch-strategy.md` を参照する

例:

- `feature/source-management`
- `feature/csv-export`
- `hotfix/article-filter`
- `fix/article-filter`
- `docs/rules-update`

### Commit

- Conventional Commits を基本にする
- コミットメッセージの見出しは日本語にする
- 先頭の種別は英語でもよいが、説明部分は日本語にする
- body を書く場合も日本語にする

例:

- `feat: ソース管理 API を追加`
- `fix: 記事一覧のページネーション条件を修正`
- `docs: API ガイドラインを追加`
- `chore: GitHub Issue テンプレートを追加`

## ADR Naming

- `docs/adr/NNNN-title.md`
- 連番4桁
- タイトルは英小文字とハイフン

例:

- `0001-monorepo-structure.md`
- `0002-shared-mysql-initial-approach.md`

## Environment Variable Naming

- 大文字スネークケース
- サービス単位で意味が分かる名前

例:

- `MYSQL_DSN`
- `ARTICLE_SERVICE_URL`
- `JWT_SECRET`

## Issue / PR Naming

- Issue はフェーズや目的が見えるタイトルにする
- PR は変更意図が分かる短いタイトルにする
- 明示的な指示がない限り、PR の題名と本文は日本語にする

例:

- `[Phase 2] ソース管理 API と UI の実装`
- `ソース管理 API と UI を実装`

## Comment Writing

- コメントは日本語で記述する
- コードの逐語説明ではなく、判断理由、前提、責務境界、落とし穴を書く
- 実装とずれやすい長い仕様説明は docs に寄せ、コードコメントは短く保つ
- service 境界、失敗時処理、順序依存、SQL の意図が見えにくい箇所を優先する

例:

- `collector 側で finish を閉じるため、ingest は記事登録だけに責務を絞る`
- `source 詳細で新しい失敗を先に見せるため、started_at DESC を正本にする`
