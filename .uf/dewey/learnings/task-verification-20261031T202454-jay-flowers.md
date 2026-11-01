---
tag: task-verification
author: jay-flowers
category: gotcha
created_at: 2026-10-31T20:24:54Z
identity: task-verification-20261031T202454-jay-flowers
tier: draft
---

A task's Verify clause is the acceptance contract for its checkbox, and a green suite cannot reveal an unexecuted one. vibe-check PR #75 (ripwire-adapter): tasks.md Task 3.3 was marked [x] while its verify clause ("a synthetic empty-edge crawl test asserts the guard fires") had no corresponding test — the shipped decision-table test verified the pure predicate, every Analyze-level fixture bore dependency edges, and deleting the guard's production wiring passed the entire suite. Rule: when reviewing an implementation against tasks.md, map each Verify clause to a concrete test by name before treating the checkbox as satisfied. Predicate-level unit tests do not satisfy crawl-level or pipeline-level verify clauses, and a verify clause naming a specific scenario (empty-edge crawl) is a demand for that scenario's fixture, not a generic coverage statement.
