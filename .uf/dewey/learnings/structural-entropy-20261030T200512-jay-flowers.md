---
tag: structural-entropy
author: jay-flowers
category: pattern
created_at: 2026-10-30T20:05:12Z
identity: structural-entropy-20261030T200512-jay-flowers
tier: draft
---

Entropy-divisor review of a specs-only OpenSpec change: when the diff is untracked markdown under openspec/changes/ with HEAD equal to base, the Go module graph cannot change — run vibe-check analyze on a temp base worktree and on the working tree anyway and let vibe-check diff emit the (empty-delta) verdict so the approval is auditable rather than assumed (verified on ripwire-adapter: verdict APPROVE, entropyDirection stable, unreliable false, all module deltas zero). Then aim findings at the spec's structural blind spots. Recurring blind spot: vibe-check diff gates (delta-Instability >= 0.15, delta-Distance >= 0.20, delta-LCOM >= 2, new cycles) never fire for ADDED packages — added packages are informational only — so any spec introducing a new package must schedule its own absolute-ceiling check (vibe-check analyze --max-lcom / --max-instability; AD-009 LCOM4 <= 3, AD-003 I < 0.7) or the new package's cohesion and instability land unmeasured by CI.
