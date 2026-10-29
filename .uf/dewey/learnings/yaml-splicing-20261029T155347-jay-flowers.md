---
tag: yaml-splicing
author: jay-flowers
category: pattern
created_at: 2026-10-29T15:53:47Z
identity: yaml-splicing-20261029T155347-jay-flowers
tier: draft
---

When a line-based splice applies multiple independent insertions whose indices were all computed against one original line slice, collect them first and apply them in DESCENDING file-position order (sort by insertIdx descending, then insert one at a time): each higher-index insertion leaves every lower precomputed index valid, so the physical order of target structures (e.g. YAML mode lists appearing out of declaration order in hand-curated files) no longer matters. The inverse (ascending or registration order) silently corrupts anchors after the first insertion. Pair this with fail-soft guards for shapes a line-based scanner cannot structurally understand — inline flow values (`key: []`, `key: {}`) and item-less keys must produce a notice and no edit, never a guess — and lock the behavior with a committed before/after fixture pair plus a second-pass idempotency assertion (re-splice the spliced output: must report present, byte-identical, no duplicates). Canonical case: vibe-check internal/scaffold spliceMatrix (init-registers-divisor-entropy, HIGH-1 fix, commit c9fb41b).
