---
tag: reusable-structural-gate
author: jay-flowers
category: gotcha
created_at: 2026-10-28T01:29:09Z
identity: reusable-structural-gate-20261028T012909-jay-flowers
tier: draft
---

AGENTS.md's documentation obligations for user-facing changes require TWO separate issues, and tasks.md generated from a proposal tends to carry only one. The proposal's "What Changes" list surfaces the unbound-force/website docs-sync issue (constitution requirement), so it lands as a task — but the repo-local documentation gate ("MUST file a documentation issue against the current repo for user-facing changes before PR merge") lives only in AGENTS.md's behavioral rules and is routinely missed when authoring tasks. Canonical case: openspec/changes/reusable-structural-gate task 3.4 specced only the website issue; the curator spec-review flagged the missing repo issue, and the fix filed both (zero-dot-force/vibe-check#58 repo gate, unbound-force/website#338 website sync). The predecessor change ci-regression-gate had both as separate tasks (5.3 repo, 5.4 website) — use that shape. Filing mechanics that worked: search each repo by title keywords first (`gh issue list --repo <owner>/<repo> --state all --search "<keywords>"`) to rule out duplicates (search several phrasings — "structural gate", "reusable workflow", the tool name); compose the body in a temp file and use `gh issue create --body-file` (avoids shell-quoting breakage on backticks and quotes); assign with `--assignee jflowers`; verify with `gh issue view` (state OPEN, assignee set) before PR submission; file the website issue first so the repo issue can cross-link its number.
