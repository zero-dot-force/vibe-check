---
tag: structural-entropy
author: jay-flowers
category: pattern
created_at: 2026-11-01T17:29:49Z
identity: structural-entropy-20261101T172949-jay-flowers
tier: draft
---

When reviewing base→PR structural deltas with vibe-check, compare the zone (and absolute Distance) fields of both analyze snapshots, not just the diff deltas: a sub-threshold ΔDistance can cross a standing zone boundary the delta gates never fire on. Verified on vibe-check PR #75 (base f2a46dce → head 6a5e15e): the metrics package gained a dependent (Ca 4→5 via the new internal/ripwireadapter), ΔI −0.0395, ΔD +0.0395 — far below the 0.20 ΔD REQUEST CHANGES gate, verdict COMMENT — yet its standing Distance crossed 0.20 (0.1772 → 0.2167), flipping the zone from main-sequence to normal. The qualitative regression is invisible in the diff output alone. Two adjacent blind spots in delta-only gating: (1) per-PR delta gates cannot see compounding across PRs (each new dependent of a concrete core package adds ~+0.03 D); (2) a package's introducing PR is a blind window for added packages — record the added package's standing metrics in the review so a baseline exists for future deltas.
