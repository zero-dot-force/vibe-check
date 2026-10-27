---
tag: openspec
author: jay-flowers
category: pattern
created_at: 2026-10-27T17:52:49Z
identity: openspec-20261027T175249-jay-flowers
tier: draft
---

When correcting a dis-proven factual claim across a change's spec artifacts, grep the entire change directory for the claim before declaring the correction complete. In zero-dot-force/vibe-check's goreleaser-release-automation change, the permissions-model correction (caller-grants, not callee-handled) updated spec.md, design.md, and tasks.md but left proposal.md:8 ("Action pinning and permissions are handled by the reusable workflows"), the proposal's Constitution V row ("contents: write only on release job"), and checklists/release-automation.md contradicting the corrected record. A partial correction converts a uniformly-wrong-but-consistent record into an internally contradictory one, and the contradiction gets frozen at archive time — in this repo, an unreconciled contradictory record (a superseded draft Dewey learning about workflow filenames) already caused a shipped HTTP 422 defect. Surface checklist for factual corrections in a change dir: proposal.md What Changes + Constitution Alignment table, design.md Context + every Decision, every specs/**/spec.md requirement, tasks.md, plus user-facing CHANGELOG/README.
