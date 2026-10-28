---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-10-28T01:56:06Z
identity: ci-cd-20261028T015606-jay-flowers
tier: draft
---

A reusable workflow that consumes PR-scoped event context should ALSO fail closed at the callee, not only guard at the caller. Route the context (e.g. github.event.pull_request.base.sha) through a step env var and fail with ::error:: when it is empty, BEFORE any checkout uses it as a ref: a caller that triggers on non-PR events without the documented `if: github.event_name == 'pull_request'` guard then gets a red job instead of a silent vacuous green gate (empty ref → actions/checkout falls back to the event SHA → base == head → zero delta → gate passes proving nothing). The callee-side check complements the caller-side guard: the caller guard prevents the bad trigger in the documented configuration, the callee check converts an undocumented misuse from silent no-op into a visible failure. Canonical case: vibe-check opsx/reusable-structural-gate (2026-10-28), flagged as a LOW defense-in-depth advisory by the code review — the shipped workflow carries the caller-side guard (dogfood job in ci.yml) but not yet the callee-side empty-ref check.
