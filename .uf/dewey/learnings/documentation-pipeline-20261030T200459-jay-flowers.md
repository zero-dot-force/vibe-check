---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-10-30T20:04:59Z
identity: documentation-pipeline-20261030T200459-jay-flowers
tier: draft
---

Spec artifacts that track documentation follow-ups as bare issue numbers can silently collapse dual documentation gates. Vibe-Check carries two distinct docs obligations for user-facing changes: a current-repo documentation issue (AGENTS.md documentation gate) and a separate website-repo sync issue (constitution Website Documentation Sync, target unbound-force/website). A bare "#NNN" reference cannot be verified against either obligation, and an issue number adjacent to same-repo issues usually indicates the repo-local one — leaving the website-sync obligation untracked. Prior changes (goreleaser-release-automation, reusable-structural-gate) established the convention of qualifying website issues explicitly (e.g., unbound-force/website#336, or a spec requirement that the issue exists). Reference follow-up issues with explicit repository qualifiers and enumerate each documentation gate separately in the spec's follow-up obligations.
