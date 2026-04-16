# AGENT.md

## Read When

- 画面表示、UI 文言、スタイル、フォーム、検索条件を変えるとき
- gateway のレスポンス変更に frontend が追従するとき

## Open First

- `src/ui/`: 対象画面
- `src/lib/api.ts`: API 呼び出し
- `src/store/searchStore.ts`: 検索条件と一覧状態
- `src/styles.css`: 全体スタイル

## Usually Skip

- backend の詳細実装
- infra / k8s docs

ただし API 契約変更時は `api-gateway` 側を確認します。

## Minimum Verification

- `cd frontend && npm run build`
