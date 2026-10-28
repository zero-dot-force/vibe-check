---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-10-28T12:14:45Z
identity: content-pipeline-20261028T121445-jay-flowers
tier: draft
---

A MEDIUM content-pipeline advisory that the author defers (\"post-merge follow-up candidate\") does not self-execute: on any re-review of the same change, re-verify the issue tracker directly (keyword search, not label search — label-based `gh issue list` returns empty in repos that track by title prefix with uncreated labels) and escalate per the Curator decision criteria (missing content issue MEDIUM+ → REQUEST CHANGES) instead of repeating the advisory. Content-pipeline issues (blog/tutorial) are the one documentation class no MUST rule mandates, so the Curator review is the only checkpoint; an advisory the author may decline converts the only checkpoint into no checkpoint. Filing mechanics for label-less repos: use the title-prefix convention (blog:/tutorial:/docs:) with NO `--label` flag — `gh issue create --label` fails on nonexistent labels; compose bodies in a temp file with `--body-file`. Canonical case: vibe-check opsx/reusable-structural-gate (2026-10-28) — the first code review flagged the missing blog issue for the flagship reusable structural gate as a MEDIUM advisory; the retrospective learning recorded it as \"a post-merge follow-up candidate rather than a filed issue\"; the same-day re-review verified the tracker still empty (open blog #27 scoped to the divisor-entropy agent; closed announcement #19 predates the capability) and escalated to REQUEST CHANGES with the full `gh issue create` command embedded in the recommendation, because the review context prohibited issue creation (graceful-degradation path carried the fix).
