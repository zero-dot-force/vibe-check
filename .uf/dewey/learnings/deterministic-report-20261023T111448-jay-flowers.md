---
tag: deterministic-report
author: jay-flowers
category: pattern
created_at: 2026-10-23T11:14:48Z
identity: deterministic-report-20261023T111448-jay-flowers
tier: draft
---

The deterministic-report pipeline run (opsx/deterministic-report, completed 2026-10-23) validated a critical workflow pattern: a change that computes architectural metrics must itself pass those metrics' gates. Iteration 1 of code review caught the new report engine raising metrics package LCOM4 from 3 to 5 — exactly the protected ΔLCOM≥2 threshold — because the review council's Divisor-Entropy persona actually RAN `vibe-check diff --gate` rather than inspecting code. The fix (moving ~2,100 lines into a new metrics/report subpackage, import path github.com/zero-dot-force/vibe-check/metrics/report, exported names stable, per-site //nolint:revive waivers for the Report-prefix stutter check) dropped the delta to zero and improved instability, converting the gate verdict from REQUEST_CHANGES to COMMENT with exit 0. Key lessons: (1) always run the project's own entropy gate on base-vs-head BEFORE declaring implementation complete — no task in tasks.md had done so; (2) untracked-file git mechanics matter — the unanchored `vibe-check` .gitignore pattern silently hid every new file under cmd/vibe-check/ from git status until anchored to `/vibe-check`; (3) golden files generated via a single-writer -update flag (with comparison tests skipping under -update) eliminate write/read races; (4) spec review iterations that surface HIGH findings early (supersession contradictions, mode-content contradictions) are cheap to fix at spec stage and would have been expensive after golden files were authored; (5) delegation to cobalt-crush-dev per-task with the orchestrator marking tasks.md checkboxes keeps bookkeeping single-writer and clean.
