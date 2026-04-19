# CI / テストレーン構成図

この図は、GitHub Actions の test-related lanes に絞って、`changes` job の判定結果が `verify`、`integration`、`coverage`、`smoke` にどう影響するかを示します。
通常のコード変更、docs-only 変更、E2E relevant change の違いを、repo 実装に即して把握できるようにしています。

```mermaid
flowchart TD
  subgraph Triggers
    pr[pull_request]
    push[push main]
    dispatch[workflow_dispatch]
  end

  subgraph Detection
    changes[changes job<br/>detect docs_only and e2e_relevant]
  end

  subgraph Verify[verify]
    verifyGate{docs_only?}
    verifyRun[make verify]
    verifySkip[skip heavy verification<br/>lightweight success]
  end

  subgraph Integration[integration]
    integrationGate{docs_only?}
    mysql[(MySQL service container)]
    articleInt[make test-article-integration]
    notificationInt[make test-notification-integration]
    integrationSkip[skip integration]
  end

  subgraph Coverage[coverage]
    coverageGate{docs_only?}
    coverageRun[make test-frontend-coverage]
    coverageArtifact[upload frontend coverage artifact]
    coverageSkip[skip coverage]
  end

  subgraph Smoke[smoke]
    smokeGate{docs_only != true and<br/>push main / workflow_dispatch /<br/>PR with e2e_relevant = true?}
    env[cp .env.example .env]
    browsers[install Playwright browsers]
    e2eUp[make e2e-up]
    e2eSeed[make e2e-seed]
    e2eRun[make test-e2e]
    report[upload Playwright report<br/>on failure]
    e2eDown[make e2e-down]
    smokeSkip[skip smoke]
  end

  pr --> changes
  push --> changes
  dispatch --> changes

  changes --> verifyGate
  changes --> integrationGate
  changes --> coverageGate
  changes --> smokeGate

  verifyGate -- yes --> verifySkip
  verifyGate -- no --> verifyRun

  integrationGate -- yes --> integrationSkip
  integrationGate -- no --> mysql --> articleInt --> notificationInt

  coverageGate -- yes --> coverageSkip
  coverageGate -- no --> coverageRun --> coverageArtifact

  smokeGate -- no --> smokeSkip
  smokeGate -- yes --> env --> browsers --> e2eUp --> e2eSeed --> e2eRun
  e2eRun -->|failure only| report
  e2eRun -->|always| e2eDown
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
