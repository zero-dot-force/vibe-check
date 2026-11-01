---
tag: documentation
author: jay-flowers
category: pattern
created_at: 2026-11-01T14:32:40Z
identity: documentation-20261101T143240-jay-flowers
tier: draft
---

Adapter-specific metric semantics require a coordinated five-surface documentation sweep when they change: (1) the universal type's GoDoc — the premise source that fixture ledgers, test rationales, and adapter docs inherit (a stale "zero = could not be determined" sentinel GoDoc silently contradicted the new "zero = measured count" convention); (2) the README metric/flag tables; (3) the CHANGELOG entry; (4) the adapter package's doc.go sourcing matrix; (5) the testdata golden-notes convention ledger. Canonical case: vibe-check PR #75 TotalLines remediation (ripwire physical lines vs Go significant lines; config-only modules report an honest 0) — all five surfaces landed in one fix series with cross-consistent numbers (95/16/16/1 Python; 0/60/16/16 TS). Skipping any surface recreates the defect class: the report engine summed a field whose own GoDoc still disclaimed determination, rendering every ripwire graph as "0 lines".
