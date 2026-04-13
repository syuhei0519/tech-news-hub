# Kubernetes Guidelines

## Purpose

このドキュメントは、Docker Compose から kind、さらに Proxmox 上の実クラスタへ移行しやすくするための Kubernetes 方針を定義する。

## Migration Policy

- ローカル開発は Docker Compose を使う
- Kubernetes 検証は kind を使う
- その後 Proxmox 上の実クラスタへ移行する
- kind と実クラスタで差分が出やすい箇所は最初から切り分ける

## Baseline Resources

各サービスは以下を前提に設計する。

- Deployment
- Service
- Ingress
- ConfigMap
- Secret
- CronJob
- PersistentVolume / PersistentVolumeClaim

## Packaging Rules

- Kubernetes manifest は Helm 前提で管理する
- 環境差分は values で吸収する
- GitOps は Argo CD を前提にする
- 直書き manifest を増やしすぎない

## Environment Separation Rules

以下は values で切り替えられるようにする。

- image tag
- replica count
- ingress host
- storage class
- resource requests / limits
- secret reference
- cron schedule
- notification channel config

## Workload Rules

### frontend

- Deployment + Service
- Ingress 経由公開

### api-gateway

- Deployment + Service
- 外部公開の中心

### article-service

- Deployment + Service
- 可能な限り安定稼働を優先

### collector-service

- 初期は Deployment でもよい
- 定期収集は最終的に CronJob か専用 worker に寄せる
- kind と本番で同じ責務境界を維持する

### notification-service

- Deployment
- 将来的に queue consumer を分離可能な構成を意識する

## State Management Rules

- MySQL は PVC 前提で扱う
- Redis と RabbitMQ も永続化の必要性を段階的に判断する
- kind では簡易永続化でもよいが、本番相当では再作成前提にしない

## Ingress Rules

- ingress-nginx を前提にする
- host 名や TLS は values で差し替える
- kind 用と実クラスタ用の差分を局所化する

## Secret Rules

- Secret は manifest 直書きを避ける
- 少なくとも values 直書きと責務を分離する
- 将来的な external secret 化を見据える
- `.env` の内容をそのまま Git 管理しない

## Config Rules

- アプリ設定は環境変数ベースを維持する
- ConfigMap と Secret の境界を明確にする
- endpoint, schedule, feature toggle は ConfigMap 側を基本にする

## Resource Rules

- 初期は全サービス 1 replica を基本にする
- 32GB のミニ PC に無理のない request / limit を設定する
- 最小構成から始め、実測で調整する

目安:

- frontend: 1 replica
- api-gateway: 1 replica
- collector-service: 1 replica
- article-service: 1 replica
- notification-service: 1 replica
- MySQL: 1 instance
- Redis: 1 instance
- RabbitMQ: 1 instance
- Prometheus / Grafana / Loki: 最小構成

## Scheduling Rules

- collector の定期実行は CronJob 化しやすい形にする
- CronJob にしても article-service の可用性に影響させない
- 重複実行時の安全性を意識する

## Observability Rules

- 各 Pod に health endpoint を持たせる
- readiness / liveness probe を順次追加する
- メトリクス収集しやすい構成にする
- ログ収集は Loki を見据えて標準出力中心にする

## GitOps Rules

- Argo CD 管理しやすいディレクトリ構成にする
- 手動 apply 依存を減らす
- Helm chart と values の責務を分離する
- 環境ごとの values ファイルを増やせる構成にする

## Kind Specific Rules

- host 名、NodePort、Ingress 動作差分を吸収しやすくする
- kind ローカルストレージ依存を本番設計に持ち込まない
- まずは最小構成で動作確認を優先する

## Proxmox Cluster Rules

- 1 control plane + 1〜2 worker の小規模構成から始めてよい
- 永続化、ストレージ、バックアップを早めに検討する
- 観測基盤の入れすぎで先に疲弊しないよう段階導入する

## Delivery Rules

### Early Stage

- Compose で完結
- kind で基本動作を確認

### Mid Stage

- Helm chart 作成
- Argo CD 同期確認

### Later Stage

- 実クラスタへ移行
- PVC とバックアップ整備
- 監視強化

## Anti-Patterns

- 環境差分を manifest 直編集で吸収し続けること
- kind 専用設定を本番設計に埋め込むこと
- Secret を Git 管理すること
- すべてを最初から高可用構成にしようとすること
