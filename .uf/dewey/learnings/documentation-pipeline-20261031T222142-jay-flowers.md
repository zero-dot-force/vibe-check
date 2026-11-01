---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-10-31T22:21:42Z
identity: documentation-pipeline-20261031T222142-jay-flowers
tier: draft
---

When a PR expands a tool's capability surface, enumerate every documentation and contract surface the change makes stale before declaring documentation done, and fix them in one commit. Two grep disciplines catch the misses: (1) grep the OLD scope-claim strings repo-wide — in vibe-check PR #75 (ripwire adapter), README/CHANGELOG/AGENTS-structure were updated, but cobra analyzeCmd/rootCmd Short+Long, the main.go package doc, and the AGENTS.md overview all still claimed Go-only, because the documentation gates enumerate repo docs but not the command layer's help text (extends learning documentation-pipeline-20261031T193051); (2) grep the canonical contract artifact for every NEW serialized field — a closed JSON Schema object (additionalProperties: false) silently rejects fields the Go model gains when the repo's own validator is hand-rolled and lenient: provenance.sourceTool* was emitted but undeclared in modulegraph.schema.json's closed input object, tests stayed green, and the published contract drifted. Also verify no test or golden asserts the stale strings before rewriting them, and note that a feedback fix may justifiably extend to unchanged files (root.go, main.go) when the PR's behavior change is what made their text false.
