---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-10-31T22:21:44Z
identity: documentation-pipeline-20261031T222144-jay-flowers
tier: draft
---

CLI help strings are user-facing documentation and belong in the docs-coverage inventory. When vibe-check analyze gained Python/TypeScript support via the ripwire adapter (PR #75), the cobra Short/Long in analyze.go and root.go and the package doc comment in main.go still described a Go-only tool: the docs task (8.5) inventoried only AGENTS.md/CHANGELOG.md/README.md, and the narrative-sections inventory in learning documentation-pipeline-20261030T200505-jay-flowers also omitted in-code surfaces. Extend that inventory pattern: whenever a command's behavior or a tool's scope changes, enumerate cobra Short/Long, the root-command scope strings, and package doc comments in the spec's docs task alongside the narrative sections. A --help surface that contradicts the README violates content.md VB-004 exactly as a stale narrative section does — and it is the first surface a CLI user consults.
