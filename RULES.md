# RULES.md

## Goal

このリポジトリでは、個人開発と学習を両立する情報収集アプリを構築する。

守るべき主目的は次の 2 点。

- 技術情報収集を効率化すること
- マイクロサービス、Kubernetes、CI/CD、監視運用を段階的に学ぶこと

## Product Rules

- 生成 AI の要約・分類は初期スコープに入れない
- 初期価値は検索、フィルタ、CSV 出力で出す
- 初期は単一ユーザー前提で設計する
- ただし将来拡張しやすい責務分離は維持する
- article-service が生きていれば記事閲覧は継続できる構成を優先する
- collector-service や notification-service 停止時も既存記事閲覧を止めない

## Scope Rules

### Included In Initial Scope

- frontend
- api-gateway
- collector-service
- article-service
- notification-service
- MySQL
- Redis
- RabbitMQ
- Docker Compose
- kind
- Helm
- GitHub Actions
- CSV 出力
- 最低限の監視

### Excluded From Initial Scope

- 生成 AI 要約
- 自動分類
- OAuth ログイン
- 複数ユーザー対応
- 高度な権限制御
- モバイルアプリ
- 複雑な推薦機能

## Architecture Rules

- マイクロサービスとして責務を分離する
- フロントからの入口は api-gateway に集約する
- サービス間の内部通信用 API は外部公開 API と分離する
- article-service を中核サービスとして扱う
- collector-service は取得と正規化に責務を限定する
- notification-service は通知判定と配送に責務を限定する
- frontend は表示責務に集中し、集計や業務判断は backend 側に寄せる

## Service Rules

### frontend

- React + TypeScript + Vite を使う
- 画面責務に専念する
- API 直接呼び出しは api-gateway に限定する

### api-gateway

- フロント向けの単一入口とする
- JWT 検証、共通ヘッダ、共通エラー制御、ログ、トレースID付与を担う
- CSV ダウンロード制御は gateway 側に寄せる

### collector-service

- RSS / API / HTML の取得アダプタを追加可能な構造にする
- robots.txt や利用規約に反しない範囲で扱う
- 外部取得結果は正規化して article-service に登録依頼する
- 取得失敗記録、リトライ、将来の非同期イベント発行を考慮する

### article-service

- 記事、ソース、ジョブ履歴の責務を持つ
- 記事一覧、詳細、検索、フィルタ、ソート、CSV、ダッシュボード集計を提供する
- 既読、お気に入り、タグ、カテゴリを管理する

### notification-service

- 新着通知、取得失敗通知、日次 / 週次ダイジェストを担当する
- 初期は画面内通知またはログ出力に留める
- Slack / メールは後続フェーズで追加する

## Data Rules

- DB は MySQL を使う
- 初期は MySQL 1 インスタンスでよい
- ただしコード上はサービスごとのアクセス責務を明確に分ける
- 重複防止には `dedupe_key` を使う
- 取得失敗はジョブ履歴やソース状態に残す
- 日時は UTC ベースで保存する
- CSV 出力の並び順と日時フォーマットは統一する
- UTF-8 出力を基本とし、Excel 利用を意識したエスケープを行う
- CSV には件数上限を設ける

## API Rules

- OpenAPI を前提に設計する
- 公開 API は `/api/v1/...` に置く
- サービス間 API は `/internal/...` に分ける
- 検索、フィルタ、ソート、ページネーションを API レベルで扱う
- エラーレスポンス形式はサービス間でできるだけ揃える

## Development Rules

- Docker Compose で全体起動できることを優先する
- ローカル環境でも複数サービスの責務分離を崩さない
- 過剰設計より学習しやすさと保守しやすさを優先する
- ただし後から kind と実クラスタへ移行しやすい境界は残す
- 変更時はドキュメントも更新する
- 作業内容に応じたドキュメント更新対象は実装者が自分で判断し、同じ変更内で反映する
- API、DB、手順、現在地、優先順位、既知課題、運用ルールの変更は、対応する docs を未更新のまま完了扱いにしない
- `.env` はコミットしない
- `.env.example` は常に最新に保つ
- Git のコミットメッセージは日本語で記述する
- 明示的な指示がない限り `main` へ直接 push しない
- 基本運用は feature branch + Pull Request とする
- release branch は明示的な指示がない限り切らない
- バグ修正は `hotfix` として扱う
- merge 方式は `merge commit` を使う
- マージ済み branch は削除してよい

## Phase Rules

### Phase 1

- article-service
- collector-service
- api-gateway
- frontend
- MySQL 接続
- 記事一覧 / 詳細 / 検索

### Phase 2

- ソース管理
- 既読 / お気に入り
- CSV 出力
- ジョブ履歴

### Phase 3

- notification-service
- RabbitMQ
- 日次 / 週次ダイジェスト
- 取得失敗通知

### Phase 4

- Docker Compose 整備
- GitHub Actions
- OpenAPI 整備

### Phase 5

- kind デプロイ
- Helm chart 作成
- Argo CD

### Phase 6

- Proxmox 実クラスタ
- 永続化
- 監視強化
- 運用改善

## Kubernetes Rules

- kind で検証できる構成を維持する
- Proxmox 上の実クラスタへ移行しやすい構成を意識する
- Deployment, Service, Ingress, ConfigMap, Secret, CronJob, PVC は早い段階から分離して考える
- 環境差分は Helm values で吸収する
- Ingress, Domain, StorageClass, Secret 管理は環境依存を分離する
- ストレージをローカル環境依存にしすぎない

## Operations Rules

- ログとメトリクスで障害追跡できるようにする
- 最初は最小構成で運用し、必要に応じて広げる
- 32GB メモリのミニ PC を前提に無理のない構成にする
- まずは全サービス 1 replica を基本とする
- MySQL, Redis, RabbitMQ, Prometheus, Grafana, Loki も最小構成から始める

## Quality Rules

- 新機能追加時は責務境界を壊していないか確認する
- article-service 停止以外で閲覧機能が致命停止しない構成を維持する
- collector の失敗が記事閲覧機能に波及しないようにする
- 将来の非同期化を前提に、処理境界を意識して実装する
- 検索と CSV のユースケースを常に優先度高く扱う

## Decision Policy

実装方針に迷った場合は次の順で優先する。

1. 学習効果が高いこと
2. Docker Compose で扱いやすいこと
3. kind / Helm / Argo CD へ移しやすいこと
4. 小規模運用で無理がないこと
5. 過剰に複雑でないこと
