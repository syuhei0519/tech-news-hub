# Known Issues

## Purpose

このドキュメントは、既知の制約や未実装事項を明文化し、新規セッションで誤解を減らすためのもの。

## Current Known Issues

### Authentication Is Not Implemented

- api-gateway に JWT の導入ポイントはあるが、実装は未着手
- 初期は単一ユーザー前提のため後ろ倒し

### Kubernetes Delivery Is Not Implemented Yet

- `deployments/k8s` はプレースホルダのみ
- kind / Helm / Argo CD は未着手

## Usage Rules

- 既知制約を見つけたらここに追加する
- 既知課題が解消されたら削除または更新する
- PR で「未対応だが既知」のものはここに記載する
