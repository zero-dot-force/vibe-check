---
tag: convention-pack-authoring
author: jay-flowers
category: gotcha
created_at: 2026-09-01T21:05:45Z
identity: convention-pack-authoring-20260901T210545-jay-flowers
tier: draft
---

When creating convention pack rules that reference vibe-check CLI flags, always verify which flags actually exist in cmd/vibe-check/analyze.go before writing the spec. In the agent-design-pack session, three spec review agents independently flagged AD-002's --max-ce and AD-009's --min-cohesion=0.5 as nonexistent or semantically incompatible with the existing CLI. The existing flags are --max-instability, --max-distance, --max-lcom (integer), --no-circular-deps, --timeout, and --output/-o. Rules referencing unimplemented flags must be explicitly marked as "forward reference to a planned feature" with a fallback enforcement strategy (e.g., heuristic review or JSON output inspection). This pattern was established in the AD-002/AD-008 fixes and should be followed for all future convention pack rules that reference planned tooling.
