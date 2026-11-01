---
tag: uf-unleash
author: jay-flowers
category: pattern
created_at: 2026-11-01T01:58:38Z
identity: uf-unleash-20261101T015838-jay-flowers
tier: draft
---

Verified on vibe-check opsx/init-scaffolds-structural-gate-caller (2026-11-01): a /uf.unleash pipeline interrupted by a session timeout resumed cleanly at the code-review phase because every earlier phase's completion lived in durable state ahead of any session memory — get_progress checkpoint steps, all 19 tasks.md checkboxes marked [x], and two committed feature-branch commits. The resume re-anchored via get_progress + parse_tasks + git status (the session-resume guard) instead of trusting the interruption summary, skipped the settled spec review and completed implement phases entirely, and the code-review council then returned APPROVE WITH ADVISORIES on the first iteration (7/7 reviewers succeeded, 3 MEDIUM advisories, no fix loop entered). Two repeatable signals: (1) when the spec-review decision gate has already applied every council decision to the artifacts before implementation begins (here, all 8 decisions including a maintainer-approved two-pin amendment), the downstream code review tends to find only verification-clause and documentation-surface advisories rather than structural defects — spec hardening pays for itself at code review; (2) MEDIUM test-quality advisories clustered exactly where tasks.md Verify clauses enumerated some but not all of a requirement's invariants (unasserted next-steps elements, untested write-ordering invariant), so Verify clauses should enumerate every testable invariant of the behavior they gate, not a representative sample.
