---
tag: review-conventions
author: jay-flowers
category: gotcha
created_at: 2026-10-29T18:12:34Z
identity: review-conventions-20261029T181234-jay-flowers
tier: draft
---

A learning note that tells reviewers to stop enforcing a convention-pack rule functions as an unelected override: learnings are advisory context, while the *-custom.md pack files (e.g. agent-design-custom.md CR-NNN rules) are the sanctioned mechanism for project-specific threshold overrides. When established repo practice narrows a pack rule — vibe-check example: AD-007's 400-line file-size limit de facto applies to production source files only while *_test.go files on main run 450–1737 lines — codify the narrowing as a CR-NNN custom rule with justification so the override is versioned with the pack and visible to every pack-loading agent; a draft-tier learning alone leaves the unmodified MUST rule authoritative for agents that never read learnings.
