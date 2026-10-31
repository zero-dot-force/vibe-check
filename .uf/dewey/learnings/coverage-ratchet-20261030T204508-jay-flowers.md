---
tag: coverage-ratchet
author: jay-flowers
category: gotcha
created_at: 2026-10-30T20:45:08Z
identity: coverage-ratchet-20261030T204508-jay-flowers
tier: draft
---

A tighten-only coverage-ratchet baseline generated on a feature branch that lags main can silently under-tighten for packages whose coverage differs between the baseline-generation tree and the tree CI actually gates: the floor records the generation-time measurement, floors only fail downward, and the gap surfaces only as an inaccuracy in the CHANGELOG's floor-provenance exceptions list. Verified on vibe-check opsx/ci-coverage-ratchet-floors: internal/scaffold floor 87.8 (truncated measurement at task-3.1 baseline generation on the branch tree) vs 91.6 measured on the final gated tree — main's post-branch-point init-registers-divisor-entropy commits added scaffold registration code and tests the baseline run never measured — a 3.8pp gap absent from the CHANGELOG's exceptions even though the cmd/coverage-ratchet and internal/coverratchet floors were reconciled in the same iteration. Reconciliation rule: compare every committed floor against the measurement on the merge result (base + PR), not just the PR head, and close gaps with targeted reviewed edits — never a blanket --update, which clobbers documented accommodation floors like metrics 82.0. Timing constraint: when only the merged tree measures higher, the PR cannot record the higher floor without failing its own head CI (the head measures the branch-tree value); either rebase onto main and re-derive the affected floors before merge, or merge as-is and raise the lagging floors on main in a tracked, reviewed follow-up.
