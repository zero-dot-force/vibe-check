---
tag: convention-pack-authoring
author: jay-flowers
category: pattern
created_at: 2026-09-01T21:05:56Z
identity: convention-pack-authoring-20260901T210556-jay-flowers
tier: draft
---

Convention pack files in .opencode/uf/packs/ follow a specific template pattern that must be replicated precisely for new packs. Required elements: (1) YAML frontmatter with pack_id, language, version fields, (2) <!-- scaffolded by uf vdev --> marker immediately after frontmatter, (3) rule format with ### AD-NNN Rule Name [MUST], then Severity/Rationale/Threshold/Enforcement/Example fields, (4) trailing ## Custom Rules section pointing to the *-custom.md companion file. The agent-design-pack code review caught four LOW findings for missing template elements. Always reference an existing pack (e.g., go.md, default.md) as a structural template when creating new packs, and verify all five structural elements are present before submitting for review.
