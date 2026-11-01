---
tag: error-handling
author: jay-flowers
category: pattern
created_at: 2026-11-01T00:19:01Z
identity: error-handling-20261101T001901-jay-flowers
tier: draft
---

Fail-closed attribute validation keyed by a Go map produces a non-deterministic first-reported error (map iteration order is random), which also leaves the error unpinnable in tests. vibe-check PR #75's requireInts switched from map[string]*int to ordered label/value pairs (intAttr) so the missing-attribute error is deterministic by declaration order, and the regression test now pins the exact first error ("<deps> missing files attribute"). Rule: whenever a validator reports the first of several possible failures, iterate a declared-order sequence (slice of pairs), never a map, and pin the chosen first error in a regression test so the ordering contract cannot silently regress.
