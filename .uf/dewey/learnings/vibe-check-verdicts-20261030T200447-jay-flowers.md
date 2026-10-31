---
tag: vibe-check-verdicts
author: jay-flowers
category: gotcha
created_at: 2026-10-30T20:04:47Z
identity: vibe-check-verdicts-20261030T200447-jay-flowers
tier: draft
---

vibe-check verdict semantics: metrics.ComputeDelta marks any graph with Status != complete (or a load-error warning) Unreliable, and DecideVerdict then forces COMMENT — never APPROVE or REQUEST_CHANGES. Any adapter spec that mandates partial status for every produced graph (e.g., approximation-warning-carrying adapters like the ripwire adapter) implicitly disables the protected entropy/new-cycle REQUEST_CHANGES gates for that language. Adapter specifications must state the verdict-layer consequence of partial-by-construction graphs: 'diff works unchanged' is true at the code level while neutralizing the gate. Found in the ripwire-adapter spec review (2026-10-30) by tracing spec.md's unconditional error-budget warnings + partial-status mandate through delta.go graphUnreliable and verdict.go's Unreliable precedence.
