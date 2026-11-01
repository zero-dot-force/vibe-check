---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-10-31T20:44:57Z
identity: test-quality-20261031T204457-jay-flowers
tier: draft
---

PR #75 ripwire-adapter TotalLines feedback (divisor-testing assessment): a fixture-ledger convention is only as strong as its premise, and an internal ledger is not a disclosure surface. golden-notes.md C3 pinned totalLines=0 and explicitly rejected summing cc.json loc — but its premise ("TotalLines exists solely to compute duplication percentage") was inherited from the metrics GoDoc and is falsified by grepping the field's consumers: metrics/report/render.go sums TotalLines for the report header's "N packages, M lines" stat and sorts modules by it, so every ripwire-analyzed project renders as "0 lines". Meanwhile the adapter's public disclosure surfaces (doc.go sourcing matrix, warning stream) omitted the field entirely, so the convention was undisclosed to anyone not reading testdata ledgers. Remediation inventory rule for any single-module-field change: the pin apparatus is triple — goldens (reflect.DeepEqual on modules), integration wantCodes lists (DeepEqual on warning codes), and the convention ledger — plus the invariant-test rationale string, which must name the field it pins and that field's own disclosure (the "(duplication unsupported)" rationale on a TotalLines pin conflated two independent fields). Reusing an existing warning code with an extended message avoids code-list churn but leaves the disclosure unguarded because goldens pin codes, not messages; a message-level assertion is then mandatory.
