# ADR 0001: Monorepo And Service Boundaries

## Status

accepted

## Context

このプロジェクトは、個人開発と学習を両立する必要がある。
一方で、マイクロサービス、API Gateway、Kubernetes、GitOps を学ぶためには、ある程度の責務分離も必要になる。

## Decision

- モノレポ構成を採用する
- `frontend`, `api-gateway`, `article-service`, `collector-service`, `notification-service` を分離する
- 記事閲覧継続性を重視し、`article-service` を中核サービスとする
- ローカルでは Docker Compose、Kubernetes 検証では kind、将来運用では Proxmox 上の Kubernetes を前提とする

## Consequences

良い影響:

- 学習コストと運用コストのバランスがよい
- 1リポジトリで docs、services、deployments を同期しやすい
- 小規模構成でもマイクロサービス設計を体験できる

悪い影響 / トレードオフ:

- 単一リポジトリ内で責務が混ざりやすい
- 初期は共有 DB など理想から簡略化する部分がある

## Alternatives Considered

- 完全モノリス:
  学習対象が狭くなり、Kubernetes / service separation の経験値が落ちる
- サービスごとの別リポジトリ:
  初期運用負荷が高く、個人開発には重い

## Notes

責務のぶれを防ぐため、`RULES.md` と各 guideline を併用する。
