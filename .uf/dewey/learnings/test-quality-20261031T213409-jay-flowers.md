---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-10-31T21:34:09Z
identity: test-quality-20261031T213409-jay-flowers
tier: draft
---

A codes-only golden comparison is not evidence that warning prose is unpinned — the pin apparatus is layered, and producer-level unit tests may hold the message assertions. vibe-check PR #75 (ripwireadapter, head aeca31c): the golden test projects Warnings to code-only lists and golden.json stores message=code placeholders by documented convention (golden-notes.md C5), which reads as "only the codes are pinned" — yet seven of the package's eleven warning codes have producer-level substring assertions pinning load-bearing message content (budget_test.go: 1.9%/2.2%/0.16% error budgets per task 4.5; denylist_test.go: pruned-directory listings; guard_test.go: both counts; degrade_test.go: parse-error paths; cycles_test.go: corroboration clause; tsabstractness_test.go: bias keywords; assembly_test.go TestDepsWindowWarning: window counts). Only directory-as-module, name-resolved-edges, ambiguity-rate, and unsupported-metric were prose-unpinned. Assessment rule for "message text can drift silently" feedback: inventory the producer tests for every warning code before accepting or rebutting the claim — and prefer substring pins that bind numbers to their phenomenon ("1.9% of edges lost"), since bare-magnitude pins survive transposition.
