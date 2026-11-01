---
tag: structural-entropy
author: jay-flowers
category: pattern
created_at: 2026-10-31T20:09:49Z
identity: structural-entropy-20261031T200949-jay-flowers
tier: draft
---

Adapter-PR entropy shape in layered metrics tools (verified on vibe-check PR #75, base 814053c → head aeca31cf): wiring a new leaf language adapter into the CLI moves exactly two existing packages, both sub-threshold. The CLI package (package main, Ca=0) gains +1 Ce with Instability pinned at 1.0 — pure fan-out, no gate relevance. The universal-model package gains +1 Ca; because it sits below the main sequence (A≈0.033, I≈0.79), the extra dependent LOWERS Instability (−0.0395, reads as "more stable") while RAISING Distance (+0.0395, 0.1772→0.2167) and flipping the zone label from main-sequence to normal. Pattern: for a concrete Layer-1 model that every adapter must import, each new dependent deepens zone-of-pain drift ≈ Ce/(Ca·(Ca+Ce)) ≈ 0.04 per adapter at the current shape; per-PR delta gates (ΔD ≥ 0.20) never fire on this slow accumulation, so entropy reviews of adapter PRs should watch the zone label and cumulative Distance, not just the verdict. Added packages themselves never gate (ripwireadapter shipped at Ca 1, Ce 16, I 0.9412, LCOM 2 — informational only), so the entire feedback cluster inside the new package (stale wiring GoDoc, dead countSourceFiles) is entropy-invisible by design.
