---
tag: structural-entropy
author: jay-flowers
category: pattern
created_at: 2026-10-30T17:33:14Z
identity: structural-entropy-20261030T173314-jay-flowers
tier: draft
---

Additive Go PRs that introduce only new packages — a leaf cmd/* CLI plus an internal/ engine with no intra-module imports — produce a nil base→PR structural delta in vibe-check: every pre-existing package holds Ca/Ce/Instability/Abstractness/Distance/LCOM exactly, added packages are informational-only in `vibe-check diff` (never gate), and the verdict is APPROVE with entropy direction stable. Verified on opsx/ci-coverage-ratchet-floors: cmd/coverage-ratchet (Ca 0, Ce 9, I 1.0 — leaf CLI shaped like cmd/vibe-check) and internal/coverratchet (Ca 1, Ce 8, I 0.89, D 0.11, LCOM 2 — inside the existing internal-family band of goadapter/scaffold), five pre-existing packages at Δ=0, zero cycles in both graphs. Companion to the same-package-growth learning (structural-entropy-20261029T155410-jay-flowers): the entropy gate bounds drift of existing structure, so acyclic, layering-consistent package growth is structurally free. Measurement caveat: when all PR work is uncommitted on a branch whose commit base is an ancestor of the recorded base ref, non-gated snapshot fields (totalLines) pick up base-range artifacts (e.g. internal/scaffold 740 base vs 206 PR from a metric-neutral main commit) while the gated Martin-metric columns stay valid — confirm by checking that per-package deltas are zero for every package outside the PR's touch set.
