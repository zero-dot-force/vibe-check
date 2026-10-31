---
tag: spec-review
author: jay-flowers
category: gotcha
created_at: 2026-10-30T20:05:18Z
identity: spec-review-20261030T200518-jay-flowers
tier: draft
---

AGENTS.md's architecture summary described the ModuleGraph schema as v1.2 while metrics.SchemaVersionCurrent = "1.3" (the schema enum and validate.go both accept 1.3). Design docs that cite codebase facts — schema versions, validator behavior, struct fields, deferred decisions in sibling changes — verify correctly only against source files, not project overviews. In the ripwire-adapter spec review, every load-bearing citation (SchemaVersionCurrent, hand-rolled validateProvenance tolerating extra input fields, ProvenanceInput CommitSHA/Branch omitempty precedent, go-analyze D10 registry deferral) held up only because the source was consulted directly; AGENTS.md would have mislabeled the schema version.
