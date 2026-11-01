---
tag: structural-entropy
author: jay-flowers
category: gotcha
created_at: 2026-10-31T22:37:34Z
identity: structural-entropy-20261031T223734-jay-flowers
tier: draft
---

Feedback-triage gotcha (vibe-check PR #75, review item #15): a claim that a new package is "unmeasured by CI entropy ceilings" self-refuted on inspection — the cited metrics (Ca=1, Ce=16, I=0.9412, LCOM=2) were themselves produced by the ecosystem's own entropy measurement (divisor-entropy drift tracking). vibe-check's CI has four measurement surfaces — structural-gate.yml whole-repo analyze (head+base snapshots), diff --gate protected delta thresholds, coverage-ratchet floors, and divisor-entropy Dewey drift tracking — but zero per-package absolute ceilings, by documented design (reusable_structural_gate.yml Analyze-head step, design.md D11: "No threshold flags are passed to analyze — gate enforcement belongs solely to diff --gate"; the analyze --max-* ceiling flags exist in the tool but no CI path passes them). Net property: newly added packages are measured but never gate on absolute metrics — delta gates police change, not state. Triage rule for future feedback cycles: verify "unmeasured"/"ceiling" claims against the workflow files and against the provenance of the reviewer's own figures before accepting; ceiling-style enforcement would reverse a documented design decision and requires spec work (threshold design, AD-003 leaf-exemption policy — the internal-adapter profile at I≈0.94 would fail a naive I<0.7 ceiling), not a feedback-loop fix.
