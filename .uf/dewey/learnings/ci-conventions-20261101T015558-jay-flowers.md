---
tag: ci-conventions
author: jay-flowers
category: pattern
created_at: 2026-11-01T01:55:58Z
identity: ci-conventions-20261101T015558-jay-flowers
tier: draft
---

A design that records a deviation note for one convention-pack MUST can still ship a departure from a different MUST in the same pack, because spec reviews audit the rules the design cites rather than the full pack. Canonical case: vibe-check opsx/init-scaffolds-structural-gate-caller (2026-10-31) — the design cites CI-001 (pinning), CI-012 (concurrency), and CI-021 (permissions) and carries a maintainer-approved deviation note for CI-001 via the D3 amendment, yet D5 names the emitted caller workflow `vibe-check-structural-gate.yaml` without ever citing CI-010, whose consumer clause requires the `ci_` filename prefix — even though the immediately preceding change (reusable-structural-gate design D8) had recorded its own naming decision explicitly against CI-010. Fix pattern: when a change authors or emits GitHub Actions workflow files, grep the active ci.md pack for every MUST and disposition each against the artifact — filename prefix, header comment, permissions placement, pinning — recording either compliance or a deviation note with the consumer upgrade path.
