---
tag: spec-verification
author: jay-flowers
category: pattern
created_at: 2026-10-31T20:24:56Z
identity: spec-verification-20261031T202456-jay-flowers
tier: draft
---

A task's Verify clause is an acceptance check at a named level, and a lower-level test does not discharge it. In vibe-check PR #75 (ripwire-adapter), Task 3.3's verify clause read "a synthetic empty-edge crawl test asserts the guard fires"; it was certified complete by guard_test.go's decision table — a predicate-level test whose case name ("zero edges over source files fires") echoed the clause's condition — while no test constructed the firing condition through Analyze. Flipping the checkbox in that state certified a verification that did not exist: deleting the one-line guardWarnings append in assembly.go passed the entire suite. Rule: before marking a task [x], read the Verify clause's nouns — "crawl test", "CLI test", "two-run byte equality" name the level and shape of the required evidence. If the shipped tests sit at a lower level than the clause names (unit predicate vs end-to-end crawl), the clause is unmet: either add the test at the named level or keep the checkbox open. A unit test whose case name matches the clause's wording is the tell that a level substitution happened, not proof the clause was satisfied.
