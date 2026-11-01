---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-11-01T00:19:06Z
identity: documentation-20261101T001906-jay-flowers
tier: draft
---

When verifying a misspelling remediation or auditing a proper noun's canonical spelling, grep for every look-alike form of the token, not just the form the feedback named: one project can carry three spellings of a single tool name across artifact classes, and the look-alikes do not substring-match each other, so each variant needs its own grep pattern. Canonical case: vibe-check PR #75's external binary — canonical "ripwire" (production exec.LookPath in internal/ripwireadapter/version.go:78, README, recorded version.txt fixtures); a nine-character "riprewire" in openspec/changes/ripwire-adapter design.md and tasks.md (the origin of the round-1 help-text typo, since fixed in analyze.go); and an eight-character "ripewire" confined to internal/ripwireadapter/fake_ripwire_test.go, where the plain installFakeRipwire helper and its LookPath self-test use that name even though the file's own doc comment claims it is "the name the production adapter resolves (design.md D2)". The misspelled island is self-consistent and inert — production-path tests use the wrapper installer, which carries the canonical name — so CI stays green and only a form-by-form grep reveals the discrepancy. Rule: enumerate transposition/insertion variants of load-bearing proper nouns and grep each separately; assert a test helper's name contract against the production constant, never by round-tripping the helper against itself.
