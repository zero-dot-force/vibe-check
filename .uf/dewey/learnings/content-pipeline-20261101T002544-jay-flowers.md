---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-01T00:25:44Z
identity: content-pipeline-20261101T002544-jay-flowers
tier: draft
---

When a spec operationalizes the content-pipeline learning (blog/tutorial tracking issues are the documentation class no MUST rule mandates), it must disposition BOTH halves of the class, not just the salient blog example — and a disposition can be a recorded "need met by <existing doc>" rather than a filed issue. Canonical case: vibe-check opsx/init-scaffolds-structural-gate-caller (2026-10-31), task 5.5: it searches the issue tracker for a blog/announcement issue per the 2026-10-28 retrospective learning but never searches for or dispositiones a tutorial-class issue, even though the change introduces a new adoption workflow (scaffold the caller, commit it, mark the check required in branch protection, uncomment tighten-only thresholds). The tutorial need happened to be satisfied by the existing README "## CI integration" section plus the command's CI next-steps output, but that coverage was incidental, not recorded. Fix pattern: content-pipeline verification tasks should enumerate all three Curator classes — docs issue, blog issue, tutorial issue — each with either a filed issue or a one-line recorded disposition of where the need is met.
