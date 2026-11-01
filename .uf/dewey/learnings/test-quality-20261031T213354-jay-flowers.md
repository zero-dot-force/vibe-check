---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-10-31T21:33:54Z
identity: test-quality-20261031T213354-jay-flowers
tier: draft
---

When a golden-comparison layer deliberately strips a structured field to an enum (vibe-check ripwireadapter: warning messages reduced to code lists, goldens carrying code-echo placeholders like "message": "error-budget"), the compensating content assertions must be enumerated per emitted value — and a ledger claim like "task 4.5 asserts warning text separately" must be audited against every member of the family, not one representative. In PR #75, 7 of 11 warning codes had prose substring pins in their owning test files, but the ledger's single citation (error-budget only) masked that directory-as-module, name-resolved-edges, ambiguity-rate, and unsupported-metric had none — including unsupported-metric, whose spec scenario mandates message content ("naming duplication as unavailable"). Review recipe for any adapter emitting warnings: list every warning code the package constructs, then grep the test package for a Message assertion (strings.Contains on w.Message) per code; code-list assertions (DeepEqual on extracted codes) never catch prose drift, and always-present warnings are the ones whose prose degrades unnoticed because their codes can never disappear.
