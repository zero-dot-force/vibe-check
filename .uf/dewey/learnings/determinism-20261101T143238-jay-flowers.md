---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-01T14:32:38Z
identity: determinism-20261101T143238-jay-flowers
tier: draft
---

A fail-closed validator that finds the first missing attribute by iterating a map (e.g. map[string]*int of wire-attribute labels) reports a RANDOM attribute, because Go map iteration order is nondeterministic — the error text itself becomes nondeterministic, which is a determinism bug in any codebase that mandates deterministic output and a flake source for tests that pin error prose. Fix: pair each label with its value in an ordered structure (variadic intAttr{label, value} pairs) so the first nil is reported in declaration order, then pin the first-error choice in the malformed-input table test with wantErr substring assertions. Applied in vibe-check PR #75 (internal/ripwireadapter/parse_deps.go requireInts; TestParseDepsOutput_MalformedInputs pins "<deps> missing files attribute" and "missing health files attribute" as the deterministic first errors, with comments explaining that the pre-fix map iteration reported a random attribute).
