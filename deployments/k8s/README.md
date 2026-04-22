# Kubernetes Deployment Assets

`deployments/k8s/charts/app` に app 向け umbrella Helm chart を配置しています。

## Directory Layout

```text
deployments/k8s/
└── charts/app
    ├── Chart.yaml
    ├── values.yaml
    ├── values-kind.yaml
    ├── values-proxmox.yaml
    └── templates/
        ├── _helpers.tpl
        ├── serviceaccount.yaml
        ├── frontend.yaml
        ├── api-gateway.yaml
        ├── article-service.yaml
        ├── collector-service.yaml
        └── notification-service.yaml
```

## Scope

- この chart は `frontend` / `api-gateway` / `article-service` / `collector-service` / `notification-service` をまとめて扱う app-only の umbrella chart です
- MySQL / RabbitMQ はこの issue ではインストール対象に含めません
- DB / AMQP の接続情報は既存 Secret 参照として values から受け取ります
- collector は現状どおり `Deployment` として扱い、`schedule` は CronJob 化に備えた予約値です

## Values Responsibility

### `values.yaml`

- chart の共通インターフェースを定義します
- 5 サービスの `enabled`、`image`、`service.port`、`ingress`、`resources`、`replicaCount`、probe を揃えます
- `global.dependencies` に app 間 URL と `MYSQL_DSN` / `AMQP_URL` の Secret 参照をまとめます
- `global.persistence` には Phase 6 で infra chart または external dependency に接続するときに使う差分の受け皿を置きます
- `article-service` や `notification-service` を無効化したまま依存先だけ残す場合は `global.dependencies.articleServiceUrl` / `global.dependencies.notificationServiceUrl` を外部 endpoint へ上書きします

### `values-kind.yaml`

- kind での最小検証向けの差分だけを置きます
- ローカル読み込み前提の image repository / tag / pull policy
- kind 用 ingress host
- 低めの replica / resource
- kind 側 Secret 名と storage class の差分

### `values-proxmox.yaml`

- Proxmox 実クラスタ移行を見据えた registry-aware な差分を置きます
- registry pull を前提にした image repository / tag / `global.imagePullSecrets`
- host / TLS / replica / resource の本番寄り差分
- Proxmox 側 Secret 名と storage class の差分

## Secret Boundary

- Secret の実値は values に入れません
- `global.dependencies.mysql.secretRef` は `MYSQL_DSN` を保持する Secret を指します
- `global.dependencies.rabbitmq.secretRef` は `AMQP_URL` を保持する Secret を指します
- 環境差分では Secret 名や key を上書きし、値そのものは cluster 側で供給します

## Render And Verify

```bash
helm lint deployments/k8s/charts/app
helm template tech-feed-hub deployments/k8s/charts/app \
  -f deployments/k8s/charts/app/values.yaml \
  -f deployments/k8s/charts/app/values-kind.yaml
helm template tech-feed-hub deployments/k8s/charts/app \
  -f deployments/k8s/charts/app/values.yaml \
  -f deployments/k8s/charts/app/values-proxmox.yaml
```

## Phase 6 Expected Additions

Phase 6 では次の差分を values 追加で吸収しやすい構成を維持します。

- `storageClass` の具体化
- TLS 証明書の実運用値
- `nodeSelector` / `tolerations` / `affinity`
- infra chart 側の persistence 連携
- Secret 供給方式の external secret 化
