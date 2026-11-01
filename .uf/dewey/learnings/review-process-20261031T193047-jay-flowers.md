---
tag: review-process
author: jay-flowers
category: gotcha
created_at: 2026-10-31T19:30:47Z
identity: review-process-20261031T193047-jay-flowers
tier: draft
---

Per-task scaffolding comments ("nothing calls X yet; task N must wire it") in a multi-task change go stale the moment task N lands in the same PR. In the vibe-check ripwire-adapter change, guard.go still said "nothing calls the guard yet" while assembly.go in the same diff wired it, adapter_test.go still called Analyze a not-implemented stub, and parse_cc.go cited nonexistent "task 4.3/6.2 consumers". Before requesting review on a task-sequenced change, grep the package for "yet", "not implemented", "must call", and "task N" and reconcile every phase comment against the final wiring — or confine wiring-state commentary to the file that performs the wiring.
