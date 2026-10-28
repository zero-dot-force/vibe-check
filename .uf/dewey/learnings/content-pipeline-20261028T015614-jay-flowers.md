---
tag: content-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-10-28T01:56:14Z
identity: content-pipeline-20261028T015614-jay-flowers
tier: draft
---

Proposal-authored tasks.md files carry only mandate-driven documentation issues — the constitution's website docs-sync and AGENTS.md's repo-local docs gate — but never Curator-domain content-pipeline opportunities (blog/tutorial tracking issues), because no MUST rule requires them; the Curator code review is the only checkpoint that surfaces them. For any change shipping a flagship capability (a reusable workflow, a new consumer surface, an adoption mechanism), verify the content pipeline by searching the issue tracker directly (`gh issue list --state all --search "<capability keywords>"`) rather than trusting the tasks list — a tasks list proving two docs issues says nothing about the third, announcement-shaped one. Canonical case: vibe-check opsx/reusable-structural-gate (2026-10-28) — the reusable structural gate is the flagship adoption surface of the tool, yet no blog/announcement tracking issue existed until the Curator code-review advisory (MEDIUM) flagged it; it remains a post-merge follow-up candidate rather than a filed issue.
