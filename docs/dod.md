# Definition Of Done

## Purpose

このドキュメントは、実装や変更を「完了」とみなすための最低条件を定義する。

## General Done Conditions

すべての変更で最低限満たすこと。

- 要件と `RULES.md` に反していない
- 責務分離を壊していない
- ローカルで最低限の検証ができている
- 必要なドキュメントが更新されている
- 未解決の制約や既知のリスクが明示されている

## API Change Done Conditions

- 実装が完了している
- gateway / service の責務境界が保たれている
- OpenAPI が更新されている
- バリデーションとエラーレスポンスが一貫している
- frontend 影響がある場合は追従している
- `docs/api-guidelines.md` と整合している

## DB Change Done Conditions

- スキーマ変更が反映されている
- 初期 SQL または migration が更新されている
- repository / service 層が追従している
- 重複防止、整合性、nullable の扱いが確認されている
- `docs/db-guidelines.md` と整合している

## Frontend Change Done Conditions

- 対象画面またはフローが動作する
- API 仕様と齟齬がない
- エラー時の表示が最低限ある
- 型エラーがない
- 既存フローを壊していない

## Collector Change Done Conditions

- 取得失敗時の挙動が考慮されている
- 重複登録が発生しにくい
- article-service の可用性と責務境界を壊していない
- 再試行時の安全性を意識している

## Kubernetes / Helm Change Done Conditions

- manifest / chart / values の役割が明確
- kind と実クラスタ差分を values で吸収しやすい
- Secret を Git に直書きしていない
- `docs/k8s-guidelines.md` と整合している

## Documentation Done Conditions

- 変更に対応する docs が更新されている
- ルール変更なら `RULES.md` を更新する
- 手順変更なら `README.md` または runbook を更新する
- 設計判断を固定したい変更なら ADR を追加する

## GitHub Workflow Done Conditions

- 関連 Issue がある、または不要理由が明確
- PR テンプレートの主要項目を埋められる状態である
- 検証項目が PR に記載できる

## Not Done Examples

以下は完了扱いにしない。

- 実装だけ入って docs が更新されていない
- DB を変えたのに schema / migration が未更新
- API を変えたのに OpenAPI が古い
- ローカルで未確認のまま「たぶん動く」で止める
- 既知の制約を黙ったままにする
