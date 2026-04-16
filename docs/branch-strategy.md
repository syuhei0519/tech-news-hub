# Branch Strategy

## Purpose

このドキュメントは、このリポジトリにおける Git ブランチ運用ルールを定義する。

## Basic Policy

- 明示的な指示がない限り、`main` へ直接 push してはいけない
- 基本運用は `feature branch + Pull Request` とする
- 1 Issue につき 1 branch を基本とする

## Main Branch Rules

- `main` は統合ブランチとして扱う
- `main` は常に安定状態を保つ
- 通常の作業は `main` で直接行わない
- 明示的に指示された場合のみ、`main` への直接 push を許可する

## Branch Types

### Feature Branch

新機能追加や段階的な機能実装で使用する。

例:

- `feature/source-management`
- `feature/csv-export`

### Fix Branch

通常の不具合修正で使用する。

例:

- `fix/article-filter`
- `fix/frontend-detail-layout`

### Hotfix Branch

バグ修正は `hotfix` 扱いとする。

例:

- `hotfix/article-pagination`
- `hotfix/collector-ingest-error`

補足:

- 本リポジトリでは、緊急性の有無にかかわらず、バグ修正系は `hotfix/*` を使ってよい

### Docs Branch

ドキュメント変更のみの場合に使用する。

例:

- `docs/api-guidelines-update`
- `docs/rules-operations-update`

### Chore Branch

設定、テンプレート、補助タスクなどで使用する。

例:

- `chore/github-actions`
- `chore/issue-templates`

## Release Branch Policy

- 明示的な指示がない限り、release branch は切らない
- 通常は `main` を統合先として運用する

## Pull Request Policy

- `main` へ入れる変更は基本的に Pull Request 経由とする
- PR では Issue と変更内容を対応づける
- PR の題名と本文は、明示的な指示がない限り日本語で記述する
- PR テンプレートに沿って影響範囲、検証、docs 更新有無を明記する

## Merge Policy

- merge 方式は `merge commit` を使用する
- `squash merge` は標準運用にしない
- `rebase merge` は標準運用にしない

理由:

- 履歴上で branch 単位のまとまりを残しやすくするため

## Branch Deletion Policy

- マージ済み branch は削除してよい
- 未マージ branch は削除前に必要性を確認する
- 長期間放置した branch は、Issue や current status を見て整理する

## Recommended Workflow

1. `main` を最新化する
2. Issue に対応する branch を切る
3. 実装する
4. `make verify` を実行する
5. Pull Request を作成する
6. `main` に merge commit で取り込む
7. マージ済み branch を削除する

## Naming Rules

- ブランチ名は小文字とハイフンを基本にする
- 種別プレフィックスを付ける
- 内容が分かる短い名前にする

使用する主なプレフィックス:

- `feature/`
- `fix/`
- `hotfix/`
- `docs/`
- `chore/`

## Exceptions

- 明示的な指示がある場合は、その指示を優先する
- 緊急対応で `main` へ直接対応が必要な場合も、明示的な指示が必要
