---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-10-30T17:33:37Z
identity: test-quality-20261030T173337-jay-flowers
tier: draft
---

When a CLI renders machine-readable output (JSON) and human-readable output (table/golden) from a shared result model, engine-level tests of the result model do not verify the rendering layer. Close the flag-combination matrix — {check, update} × {table, JSON} — with at least one assertion per combination: status-to-field mappings (e.g., a refused package's floor rendering as null in JSON or "-" in the table, updated:true serialization) live only in the renderer and regress silently otherwise, because the engine tests prove the statuses while nothing proves their translation. Seen in vibe-check's cmd/coverage-ratchet: TestRunRatchet_JSON ran check mode only and TestRunRatchet_Update ran table mode only, leaving report.go's refused/updated JSON branches unexecuted under test.
