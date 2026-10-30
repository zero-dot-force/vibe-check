---
tag: openspec
author: jay-flowers
category: pattern
created_at: 2026-10-30T17:33:33Z
identity: openspec-20261030T173333-jay-flowers
tier: draft
---

When a review council demands a spec gap be closed mid-iteration, the spec amendment that closes it still crosses the implementation-phase rule limiting spec-artifact edits to factual corrections. The workable pattern from vibe-check's ci-coverage-ratchet-floors fix iteration: amend the spec to match behavior that the spec-review round already advised and the code-review round already examined (never to bless unreviewed behavior), append a deviation note to the affected task recording exactly what changed and why, keep `openspec validate --strict` green, disclose the amendment in the fix-iteration walkthrough, and carry the amendment into the PR body so the human reviewer sanctions the spec change explicitly. The durable record is what converts a phase-boundary violation into a documented reconciliation.
