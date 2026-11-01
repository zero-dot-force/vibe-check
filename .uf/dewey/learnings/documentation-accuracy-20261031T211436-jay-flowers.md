---
tag: documentation-accuracy
author: jay-flowers
category: gotcha
created_at: 2026-10-31T21:14:36Z
identity: documentation-accuracy-20261031T211436-jay-flowers
tier: draft
---

Phase-status comments in test files go stale silently because no gate reconciles comment text with code state. In vibe-check PR #75 (ripwire-adapter), the adapter_test.go:8 header still claimed "Analyze is a not-implemented stub … so it carries no tests yet" and version_test.go:53 still read "// the not-implemented stub, not a pin error" after Analyze shipped and gained production-path coverage in version_test.go, assembly_test.go, and integration_test.go — a coverage reviewer reading the header would wrongly conclude Analyze was untested. Before requesting review on a package that evolved through stub phases, grep its test files for "stub", "not-implemented", and "no tests yet" and reconcile every hit against the current code; phrase comments to state the present contract ("Analyze is exercised via version_test.go"), never the schedule ("not implemented yet").
