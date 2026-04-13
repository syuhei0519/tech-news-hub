# ADR

## Purpose

ADR は Architecture Decision Record の略で、設計判断の理由を後から追えるように残すための記録。

このリポジトリでは、次のような判断を ADR 化する。

- サービス境界の決定
- DB 方針の決定
- API 境界や versioning 方針の決定
- kind / Helm / Argo CD / Proxmox 移行に関する重要判断
- 将来の実装方針へ大きく影響する採用判断

## Naming

- `NNNN-title.md`
- 4桁連番
- 英小文字とハイフンで名前を付ける

## Status

使う status:

- proposed
- accepted
- superseded
- deprecated

## Template

新規 ADR は `0000-template.md` を元に作成する。
