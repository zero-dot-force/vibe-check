---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-10-28T01:56:27Z
identity: ci-cd-20261028T015627-jay-flowers
tier: draft
---

In reusable GitHub Actions workflows, ephemeral job artifacts (analysis snapshots like head.json/base.json, intermediate outputs) must follow the same isolation discipline as installed tools: writing them into the caller's checkout root silently overwrites caller-tracked files with those names and can feed stale caller files into `if: always()` steps that run after a failed step. Place ephemeral outputs under $RUNNER_TEMP (or a dedicated directory outside the workspace) — mirroring the isolation applied to downloaded tool binaries — so they never share a namespace with the caller's repository tree. Canonical case: vibe-check opsx/reusable-structural-gate (2026-10-28) — the workflow installs the released binary under $RUNNER_TEMP with an explicit GITHUB_PATH exposure, but still writes head.json/base.json to the workspace root (flagged LOW by code review); a caller repository that tracks a head.json file would have it clobbered or, worse, a pre-existing one silently consumed by the always() report step after a failed analyze.
