# CI / テストレーン構成図

この図は、GitHub Actions の test-related lanes に絞って、`changes` job の判定結果が `verify`、`integration`、`coverage`、`smoke` にどう影響するかを示します。
通常のコード変更、docs-only 変更、E2E relevant change の違いを、repo 実装に即して把握できるようにしています。

```mermaid
flowchart TD
  subgraph Triggers[トリガ]
    pr[pull_request]
    push[push(main)]
    dispatch[workflow_dispatch]
  end

  subgraph Detection[変更判定]
    changes[changes job<br/>docs_only と e2e_relevant を判定]
  end

  subgraph Verify[verify]
    verifyGate{docs_only?}
    verifyRun[make verify]
    verifySkip[重い検証を省略<br/>軽量成功で完了]
  end

  subgraph Integration[integration]
    integrationGate{docs_only?}
    mysql[(MySQL service container)]
    articleInt[make test-article-integration]
    notificationInt[make test-notification-integration]
    integrationSkip[integration を skip]
  end

  subgraph Coverage[coverage]
    coverageGate{docs_only?}
    coverageRun[make test-frontend-coverage]
    coverageArtifact[frontend coverage artifact を upload]
    coverageSkip[coverage を skip]
  end

  subgraph Smoke[smoke]
    smokeGate{docs_only ではなく<br/>push(main) / workflow_dispatch /<br/>e2e_relevant = true の PR か?}
    env[.env.example を .env にコピー]
    browsers[Playwright browsers を導入]
    e2eUp[make e2e-up]
    e2eSeed[make e2e-seed]
    e2eRun[make test-e2e]
    report[失敗時に Playwright report を upload]
    e2eDown[make e2e-down]
    smokeSkip[smoke を skip]
  end

  pr --> changes
  push --> changes
  dispatch --> changes

  changes --> verifyGate
  changes --> integrationGate
  changes --> coverageGate
  changes --> smokeGate

  verifyGate -- はい --> verifySkip
  verifyGate -- いいえ --> verifyRun

  integrationGate -- はい --> integrationSkip
  integrationGate -- いいえ --> mysql --> articleInt --> notificationInt

  coverageGate -- はい --> coverageSkip
  coverageGate -- いいえ --> coverageRun --> coverageArtifact

  smokeGate -- いいえ --> smokeSkip
  smokeGate -- はい --> env --> browsers --> e2eUp --> e2eSeed --> e2eRun
  e2eRun -->|失敗時のみ| report
  e2eRun -->|常に後処理| e2eDown
  report --> e2eDown
```

## 補足

- `verify` は workflow 上は常に起動し、docs-only 変更では `make verify` を省略して軽量成功にします。
- `integration` は MySQL service container 付きで article-service と notification-service の integration test を分離実行します。
- `smoke` は docs-only 以外かつ `push main`、`workflow_dispatch`、または `e2e_relevant=true` の PR でだけ走ります。

## repo からは未確定

- branch protection や required checks の GitHub 設定は repo 外にあるため、この図には含めていません。

## 主な根拠ファイル

- `.github/workflows/ci.yml`
- `Makefile`
- `scripts/e2e/wait-for-stack.sh`
- `scripts/e2e/seed.sh`
- `frontend/playwright.config.ts`

## 図に含めていないもの

- `commit-message-check` と `pr-body-check` は別 workflow の PR guardrails のため、この図本体には含めていません。
