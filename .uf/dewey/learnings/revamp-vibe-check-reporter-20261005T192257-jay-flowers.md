---
tag: revamp-vibe-check-reporter
author: jay-flowers
created_at: 2026-10-05T19:22:57Z
identity: revamp-vibe-check-reporter-20261005T192257-jay-flowers
tier: draft
---

The code review phase for the revamp-vibe-check-reporter change (opsx/revamp-vibe-check-reporter) found all 4 findings in a single file: .opencode/references/vibe-check-example-report.md. The example report had drifted from the agent's actual output format contract in four ways: (1) recommendation #3 title "Reduce abstractness" didn't match its body which correctly identified the issue as 0 abstractness suggesting "Add interfaces", (2) the legend block had a "**Legend**:" prefix that the agent template does not emit, (3) a coupling note used "Stable (Ca > 10)" instead of the template's "Heavily depended-on (Ca > 10)" phrasing, and (4) the Instability Spread grade showed "C" where the threshold table says it should be "C+" for ≤ 2 extremes. This confirms a pattern: example/reference files drift from templates because they must be manually kept in sync with no automated verification. The fix was trivial (4 one-line edits) but the pattern matters: every manually-maintained reference file is a drift risk. The repo now has two such fragile sync pairs: scaffold assets ↔ .opencode source (protected by contract_test.go), and example report ↔ agent template (no automated test).
