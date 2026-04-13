# Glossary

## Purpose

このドキュメントは、プロジェクト内の用語の意味を固定するためのもの。

## Terms

### Article

収集対象から取り込まれ、一覧・検索・詳細表示の対象になる記事または更新情報。

### Source

記事を取得する外部情報源。例: Kubernetes Blog、GitHub Changelog。

### Fetch Job

collector による 1 回の取得実行単位。成功 / 失敗、取得件数、重複件数などを持つ。

### Dedupe Key

重複記事を防ぐための一意判定キー。再試行時の冪等性にも使う。

### Public API

frontend から利用される `api-gateway` 経由の API。`/api/v1/...` に配置する。

### Internal API

サービス間通信用の API。`/internal/...` に配置する。

### Gateway

frontend の単一入口。認証、共通エラー処理、レスポンス整形を担う。

### Ingest

collector-service が正規化した記事データを article-service に登録依頼する処理。

### Phase

段階的な実装順序。Phase 1 から Phase 6 まで定義済み。

### Compose

ローカル開発用の Docker Compose 環境。

### kind

ローカル Kubernetes 検証用のクラスタ環境。

### Proxmox Cluster

将来の本番相当運用を想定する、ミニPC上の Kubernetes クラスタ。
