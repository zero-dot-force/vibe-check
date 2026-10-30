---
tag: structural-entropy
author: jay-flowers
category: gotcha
created_at: 2026-10-30T17:33:17Z
identity: structural-entropy-20261030T173317-jay-flowers
tier: draft
---

vibe-check diff totalLines can move backward for packages a PR never touched when the feature branch was cut behind the review base. Verified on opsx/ci-coverage-ratchet-floors (fix iteration 2 re-review): base upstream/main a5fd0e8 vs the working tree at 3e354b7 (git merge-base HEAD a5fd0e8 = HEAD, so the branch point is an ancestor of the base) reported internal/scaffold 740→206 and cmd/vibe-check 1345→1325 totalLines, while every gated Martin metric (Ca, Ce, Instability, Abstractness, Distance, LCOM), cycles, and duplication blocks held exactly — the missing lines are main's post-branch-point additive commits (init-registers-divisor-entropy), which git restores at merge time since the PR touches disjoint files. Reviewer rule: a totalLines-only asymmetry with zero drift on the six gated metrics signals branch/base divergence, not a PR-caused structural regression; confirm with git merge-base before treating it as a finding.
