---
tag: address-feedback
author: jay-flowers
category: gotcha
created_at: 2026-10-31T22:37:57Z
identity: address-feedback-20261031T223757-jay-flowers
tier: draft
---

Feedback claims that a new package is "unmeasured by CI" must be verified against the actual workflow files before acceptance — the claim can be false on measurement while true only on mechanism-absence. Canonical case: vibe-check PR #75 review item #15 ("internal/ripwireadapter unmeasured by CI entropy ceilings"). The package was in fact measured three ways: .github/workflows/structural-gate.yml runs whole-repo `vibe-check analyze` on PR head AND base (no exclude list) before `diff --gate`; ci.yml's coverage-ratchet enforces per-package floors (ripwireadapter registered at 91.6 in coverage-floors.json); divisor-entropy drift-tracks stored analyze snapshots. What did not exist was a per-package ABSOLUTE ceiling — a pre-existing repo-wide property (AD-002 Ce<10 / AD-003 I<0.7 live in the agent-design convention pack as review-agent-enforced rules; CLI ceiling flags are documented forward references), untouched by the PR. Structural nuance: delta gates cannot fire on ADDED packages (no base counterpart), so new packages are verdict-invisible by design and agent review of absolute metrics is the compensating control. Telltale signal: feedback that quotes per-package Ca/Ce/I/LCOM values proves measurement occurred — those numbers come from the analyzer. Correct disposition: AUTHOR-DECIDES (informational reply citing the existing mechanisms, or a separately-specced tighten-only ceiling feature) — never an inline feedback fix, since CI changes require the spec-first workflow.
