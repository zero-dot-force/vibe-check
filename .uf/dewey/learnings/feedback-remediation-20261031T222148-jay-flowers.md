---
tag: feedback-remediation
author: jay-flowers
category: decision
created_at: 2026-10-31T22:21:48Z
identity: feedback-remediation-20261031T222148-jay-flowers
tier: draft
---

Feedback-fix commits may extend beyond a PR's changed-file inventory when the review feedback explicitly names an unchanged file whose content the PR's own changes invalidated. Precedent: vibe-check PR #75 Item 13 named root.go:14-17 (unchanged by the PR; rootCmd scoped the tool "for Go codebases") because the PR's multi-language routing made that scoping false — this is downstream impact of changed lines, not pre-existing staleness, and not scope creep. Three-part test: (a) the feedback names the file, (b) the text was true before the change and false after it, (c) the project's other documentation surfaces already claim the new behavior, leaving the untouched surface internally contradictory. Absent any one of the three, keep the fix inside the diff's own files.
