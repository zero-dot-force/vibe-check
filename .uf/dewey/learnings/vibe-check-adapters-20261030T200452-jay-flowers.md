---
tag: vibe-check-adapters
author: jay-flowers
category: pattern
created_at: 2026-10-30T20:04:52Z
identity: vibe-check-adapters-20261030T200452-jay-flowers
tier: draft
---

goadapter's analyzer security posture has three legs: metrics.ValidateProjectPath on the user-supplied path (Analyze Step 1, internal/goadapter/adapter.go), metrics.SanitizeEnvironment for subprocess environments, and bounded output capture. The CLI does not validate paths centrally — each adapter owns it. Specs for new metrics.Adapter implementations that claim goadapter parity must enumerate all three controls in requirements and tasks; omitting ValidateProjectPath while listing it in a constitution-alignment table produces an unverifiable security claim (found in the ripwire-adapter spec review, 2026-10-30).
