# Release And Migration Policy

## Purpose

このドキュメントは、アプリ、API、DB、Kubernetes 構成の変更を安全に進めるための方針を定義する。

## Basic Rules

- 破壊的変更は避けるか、段階的に行う
- API、DB、インフラ変更は影響範囲を明示する
- Compose での再現性を維持する
- kind と実クラスタ移行を妨げる変更を避ける

## Release Policy

### Small Increment Policy

- 小さく分けてマージする
- フェーズをまたぐ大規模変更を一度に入れない
- 機能、ドキュメント、運用変更を可能な限り対応づける

### Merge Readiness

- `docs/dod.md` を満たす
- 必要な docs が更新されている
- 主要な検証結果を PR に記載する

## API Change Policy

- 公開 API の互換性を優先する
- 変更時は OpenAPI を更新する
- frontend に影響する場合は同一 PR か関連 PR で追従する
- 破壊的変更は versioning または移行期間を設ける

## DB Migration Policy

- DB 変更はアプリ変更と整合させる
- 可能なら additive change を先に行う
- 既存カラム削除や意味変更は段階的に進める
- 初期 SQL 管理から migration 管理へ移る場合も手順を文書化する

## Config Change Policy

- 新しい環境変数を追加したら `.env.example` を更新する
- Secret を Git に含めない
- ConfigMap / Secret / values の責務を崩さない

## Compose Change Policy

- `docker compose config -q` が通ること
- ローカル起動性を優先する
- 一部サービス変更でも全体起動を壊さないようにする

## Kubernetes Change Policy

- values で環境差分を吸収できるようにする
- kind 固有設定を本番前提に埋め込まない
- PVC, Ingress, Secret, Domain, StorageClass の差分は局所化する

## Migration Types

### Code Migration

- service 間責務の変更
- internal API 契約の変更

必要なこと:

- docs 更新
- 影響範囲確認
- 段階移行可能ならその方針を記載

### Data Migration

- テーブル追加
- カラム追加
- カラム意味変更
- インデックス追加

必要なこと:

- schema 更新
- migration 更新
- rollback 観点
- データ整合性確認

### Infrastructure Migration

- Compose -> kind
- kind -> Proxmox
- static config -> managed config

必要なこと:

- runbook または手順書更新
- values 差分整理
- 検証ステップ明記

## Rollback Policy

- rollback が難しい変更は事前に PR で明示する
- DB の destructive change は後戻り方法を考える
- リリース前に feature isolation が可能か検討する

## Documentation Requirements

次の変更では docs 更新が必須。

- 新しい API
- DB schema 変更
- 新しい運用手順
- リソース配分や k8s 方針の変更
- ルールや要件の変更

## Session Handoff Policy

新規セッションへ引き継ぐために、以下を残す。

- Issue
- PR
- docs
- ADR
- TODO ではなく、判断理由を含む記録

## Anti-Patterns

- docs を後回しにして設計変更だけ入れる
- 破壊的変更を小さな修正として混ぜる
- `.env.example` を更新しない
- OpenAPI を古いまま放置する
- kind でしか動かない設定を固定する
