# AGENT.md

## Responsibility

`frontend` は React + TypeScript + Vite による UI 実装です。

- 記事一覧、詳細、検索、通知、ソース管理の画面を提供する
- バックエンド通信は `api-gateway` 経由で行う
- 画面状態と検索条件の保持を担う

## Read First

作業内容に応じて次を起点に確認します。

- `src/main.tsx`: エントリポイント
- `src/lib/api.ts`: API Gateway との通信
- `src/store/searchStore.ts`: 検索条件の状態管理
- `src/ui/`: 画面コンポーネント群
- `src/styles.css`: 全体スタイル

## Implementation Notes

- API 追加やレスポンス変更がある場合は `src/lib/api.ts` と対象画面を合わせて更新する
- 検索や一覧の挙動を変える場合は `searchStore` の状態遷移を確認する
- 画面追加時は既存の `src/ui/*Page.tsx` の粒度に揃える
