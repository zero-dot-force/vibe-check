---
tag: coverage-ratchet
author: jay-flowers
category: gotcha
created_at: 2026-10-30T20:44:59Z
identity: coverage-ratchet-20261030T204459-jay-flowers
tier: draft
---

Self-gating coverage ratchets generate their baseline on the feature branch, but CI enforces it on the post-merge tree: when main lands additive commits (code and tests) in packages the branch never touched after the branch point, the merged tree measures higher than the branch tree, leaving committed floors with silent headroom the gate cannot see (floors fail only downward). Reconciliation against "the final tree" (per the existing post-baseline headroom learning) must therefore mean the post-merge tree — merge main locally or project the merge before recording floors. When reconciling, prefer a targeted reviewed floor edit over a blanket --update so documented variance accommodations (e.g., vibe-check's metrics 82.0 for flaky subprocess coverage) survive. Canonical case: vibe-check opsx/ci-coverage-ratchet-floors — internal/scaffold floor 87.8 vs ~91.6 post-merge measurement, headroom introduced by main's init-registers-divisor-entropy commits landing after the branch point, invisible to a gate that passed at-floor on the branch tree.
