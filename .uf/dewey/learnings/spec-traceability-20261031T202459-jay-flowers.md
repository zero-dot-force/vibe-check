---
tag: spec-traceability
author: jay-flowers
category: pattern
created_at: 2026-10-31T20:24:59Z
identity: spec-traceability-20261031T202459-jay-flowers
tier: draft
---

A tasks.md verify clause that names a test level ("a synthetic empty-edge crawl test asserts the guard fires") must be satisfied at that level, not by unit tests of the underlying predicate. When implementation is split across tasks (guard predicate in task 3.3, Analyze wiring in task 6.2), the wiring task inherits the verify obligation, and the earlier task cannot be marked [x] on the predicate's decision-table test alone. During review, the reliable verification method is exhaustive enumeration of the production entry point's test callers (who invokes Analyze, with which fixtures?) — if no caller exercises the clause's named scenario, the clause is unmet even when function-level tests exist and pass.
