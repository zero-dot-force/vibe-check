---
tag: address-feedback
author: jay-flowers
category: gotcha
created_at: 2026-10-31T22:37:40Z
identity: address-feedback-20261031T223740-jay-flowers
tier: draft
---

Feedback claims that a package is "unmeasured by CI" must be verified against the repo's actual workflow files AND the delta engine's new-module semantics before acceptance. Vibe-check PR #75 feedback item: "internal/ripwireadapter unmeasured by CI entropy ceilings" — verification showed the package is measured by every mechanism the repo has: (1) structural-gate.yml runs whole-repo `vibe-check analyze` on both head and base (the reviewer's own cited Ca/Ce/I/LCOM values came from this tooling), (2) the diff report lists new packages under Added, (3) the ci.yml coverage ratchet measures the package via its coverage-floors.json registration, and (4) divisor-entropy drift tracking stores whole-repo snapshots. What the reviewer actually observed is a designed property: metrics/delta.go treats Added modules as report-only ("no Delta entry and no gate effect"), so a new package's absolute metrics never feed the delta gate at introduction time, and the repo deliberately wires no absolute --max-* ceilings into CI ("gate enforcement belongs solely to diff --gate", reusable_structural_gate.yml). "Entropy ceilings" is an entropy-sentinel concept this repo's CI never implemented for any package. Triage checklist for such claims: (a) read the workflow's actual analyze/diff invocations for path filters and threshold flags, (b) check the delta engine's handling of modules absent from the base graph, (c) distinguish "measured" (present in snapshots/reports) from "enforced" (gate-fed) — reviewers often mean the latter while saying the former, and (d) note that wiring uniform --max-instability ceilings into CI would fail established internal-adapter profiles (adapters sit at I≈0.94 by function: efferent-heavy, Ca=1 from the CLI), so ceiling enforcement needs per-profile design work, not a flag flip.
