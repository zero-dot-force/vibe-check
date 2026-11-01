---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-10-31T19:30:55Z
identity: documentation-20261031T193055-jay-flowers
tier: draft
---

Forward-referencing wiring comments rot at merge time. In vibe-check PR #75 (ripwire-adapter), guard.go shipped with the GoDoc "Wiring point: nothing calls the guard yet. Task 6.2 (Analyze graph assembly) must call it with the real parsed numbers..." while assembly.go — task 6.2, in the same PR — already calls zeroCouplingGuard and forces partial status. The merged comment misdescribes the shipped code in both directions (it claims the guard is uncalled and that the wiring is pending), and a reader trusting it could double-wire the guard. The prior review council flagged the area as "wiring unreconciled" but the stale comment survived because it reads like design rationale. Rule: when a PR lands scaffolding and its consumer in sequence, sweep the scaffolding's GoDoc for task-sequencing comments before merge — comments must describe what the code does at HEAD, never what a later task will do.
