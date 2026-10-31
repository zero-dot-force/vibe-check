---
tag: coverage-ratchet
author: jay-flowers
category: gotcha
created_at: 2026-10-30T20:45:02Z
identity: coverage-ratchet-20261030T204502-jay-flowers
tier: draft
---

A coverage-ratchet baseline committed from a feature branch that diverged from main can leave packages the PR never touched measuring higher on main than their recorded floors — branch-divergence headroom. Unlike same-PR headroom (fixable by re-recording before submission), divergence headroom cannot be reconciled inside the PR: raising the floor to main's measurement fails the gate on the PR tree, which measures the branch's value. The durable record must either enumerate the divergence alongside the other documented floor exceptions, or the team consumes it via --update immediately post-merge; a PR-body disclosure alone leaves a CHANGELOG that enumerates its floor exceptions incomplete on merge day. Case: vibe-check opsx/ci-coverage-ratchet-floors — internal/scaffold floor 87.8 (branch-tree measurement; the PR adds no scaffold files) vs 91.6 disclosed as the main-tree measurement, because main's init-registers-divisor-entropy landed after the branch point.
