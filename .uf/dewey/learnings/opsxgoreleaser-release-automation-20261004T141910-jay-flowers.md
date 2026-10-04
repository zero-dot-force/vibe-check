---
tag: opsxgoreleaser-release-automation
author: jay-flowers
created_at: 2026-10-04T14:19:10Z
identity: opsxgoreleaser-release-automation-20261004T141910-jay-flowers
tier: draft
---

When the version template variable changes ({{.Version}} vs {{.Tag}}) during spec review fix iterations, the fix must be applied across ALL four artifacts: proposal.md, design.md, spec, and tasks. In this case, spec and tasks were updated to {{.Tag}} but design.md D3 was missed, causing a MEDIUM cross-document contradiction finding from the Entropy Divisor. The Scribe persona is the best detector of this kind of inconsistency because it traces through all artifacts systematically.
