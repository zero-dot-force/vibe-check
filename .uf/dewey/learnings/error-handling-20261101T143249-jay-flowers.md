---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-01T14:32:49Z
identity: error-handling-20261101T143249-jay-flowers
tier: draft
---

Fail-closed validators that report "the first missing attribute" from a map[string]*T iterate in randomized order, so the named attribute varies run to run and no substring assertion can pin the error message. Restructure as ordered label/value pairs passed as variadic struct args (e.g. type intAttr struct{ label string; value *int }; requireInts(intAttr{"files", root.Files}, ...)): declaration order then defines a deterministic first-error that tests can pin with wantErr substrings, and a determinism bug becomes a visible regression instead of flaky error text (vibe-check ripwireadapter parse_deps.go requireInts).
