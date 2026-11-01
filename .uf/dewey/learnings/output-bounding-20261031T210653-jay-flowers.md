---
tag: output-bounding
author: jay-flowers
category: pattern
created_at: 2026-10-31T21:06:53Z
identity: output-bounding-20261031T210653-jay-flowers
tier: draft
---

In vibe-check PR 75 (ripwire-adapter), the bounded subprocess-output preview was written twice inside one package: previewVersionOutput (version.go:144, cap 120, strings.TrimSpace) and previewBytes (runner.go:151, cap 512, no trim) — 7 byte-identical consecutive body lines under the same local constant name maxPreview holding different values, and only the version copy's doc comment carried the flood-protection rationale ("a hostile or corrupt binary cannot flood the error"). It was the third copy-and-drift instance in that single PR (after the fake-ripwire test-wrapper signature drift and the %q-vs-single-quote quoting drift), confirming the pattern: same-concern helpers copied per call site drift within one changeset, and the copy lacking the security comment is the one that loses the control. Rule: the cross-binary two-copy precedent (extraction deferred to the third binary, tracked as issue #73) does not extend inside a package — same-package extraction is free, so single-source at the second call site: one helper with an explicit limit parameter (or named constants with a documented reason for each deliberately different bound), one documented trim policy, and the security rationale in the shared helper's GoDoc. Neither test file pinned the caps or trim behavior, so the extraction was test-surgery-free — drift survives precisely because nothing observes it; a small table test on the shared helper pins the policy.
