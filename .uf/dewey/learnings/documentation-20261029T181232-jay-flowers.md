---
tag: documentation
author: jay-flowers
category: pattern
created_at: 2026-10-29T18:12:32Z
identity: documentation-20261029T181232-jay-flowers
tier: draft
---

Triaging PR-declared carry-forward documentation advisories: record each declared inaccuracy as a normal finding at impact-derived severity (the review record must carry it), then verify the cited tracking issues both exist AND actually list the deferred fix before assuming coverage — verifying existence alone is not enough. Canonical case: vibe-check PR #65 (init-registers-divisor-entropy, 2026-10-29): docs issues #63 and unbound-force/website#344 existed, were open, and cross-linked, but #63's body did not name the deferred README flow-collection sentence inversion, so the finding's recommendation had to direct the fix into #63 rather than cite it as already tracked. Verdict mapping: a declared, tracked MEDIUM documentation inaccuracy (present-but-imprecise user-facing sentence describing edge-case behavior) yields APPROVE WITH ADVISORIES — the Curator gate's blocking branch fires on missing documentation or missing tracking issues, not on a deferred fix whose declaration, learning capture, and issue trail all exist.
