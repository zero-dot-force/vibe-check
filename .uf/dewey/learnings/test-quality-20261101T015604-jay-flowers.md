---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-01T01:56:04Z
identity: test-quality-20261101T015604-jay-flowers
tier: draft
---

When a spec scenario enumerates multi-element output content (\"a next-steps section covering A, B, C, D, E\"), an acceptance test that substring-asserts only the section header plus one or two elements leaves the rest unguarded — a regression deleting the unasserted elements stays green. Audit each enumerated element for a corresponding assertion, prioritizing elements that document the mitigation for a known silent-failure hazard (canonical case: vibe-check init --ci's next-steps line \"edit `branches:` when your default branch is not `main` — the gate never triggers otherwise\", the only in-output mitigation for the hardcoded branches-filter + required-check deadlock). The same rule applies to design-mandated write ordering: \"deploy X before Y so path conflicts fail before Y mutates\" is a testable invariant — pre-create the conflict (e.g., a regular file at the target directory path) and assert Y is untouched. The primary atomicity test usually covers only the validation gate, not the conflict branch, so the ordering claim otherwise rests on nothing.
