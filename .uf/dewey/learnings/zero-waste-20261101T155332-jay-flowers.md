---
tag: zero-waste
author: jay-flowers
category: pattern
created_at: 2026-11-01T15:53:32Z
identity: zero-waste-20261101T155332-jay-flowers
tier: draft
---

When the Guard reviews a test-only fix commit in a review-council fix loop, the decisive zero-waste check is behavioral traceability: trace each new test's expected values against the current implementation (not the spec's aspiration). A test whose assertions match real code paths is coverage completion — in scope for the governing change's test strategy even when the commit lands on a PR branch named for a different stacked change (stacked changes on one PR share the fix loop). A test asserting behavior the implementation does not have is orphaned code and a zero-waste violation. Complementary governance checks: verify imports the tests rely on pre-exist (appended test hunks often omit import diffs), confirm no gate values moved (adding tests is always safe under a tighten-only coverage ratchet since floors are minimums), and confirm task-bookkeeping stays untouched — review-phase fixes correctly do not edit tasks.md. Canonical case: vibe-check PR #75 fix commit 78ac86c, internal/scaffold edge-case tests verified against lines.go/splice.go/register.go/scaffold.go.
