---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-10-28T01:56:36Z
identity: ci-cd-20261028T015636-jay-flowers
tier: draft
---

CI-001 (unbound-force CI convention pack) requires every GitHub Actions uses: reference to be pinned to a full 40-character commit SHA, but a same-repository dogfood caller of a reusable workflow must use the literal local form `./.github/workflows/<file>.yml`, which cannot carry a ref — the pin is effectively the caller's own ref (the PR under test). When authoring a reusable workflow plus a same-repo dogfood consumer, record the CI-001 local-reference exemption rationale in the change's design decisions so actionlint-clean reviewers (and future readers diffing convention compliance) see the deviation is structural, not an oversight. Canonical case: vibe-check opsx/reusable-structural-gate (2026-10-28) — ci.yml's structural-gate-dogfood job uses `uses: ./.github/workflows/reusable_structural_gate.yml`, documented as design decision D8-adjacent rationale; external consumers pin by repo ref (e.g. zero-dot-force/vibe-check/.github/workflows/reusable_structural_gate.yml@<sha>), which IS sha-pinnable and that is the pin README consumers are told to use.
