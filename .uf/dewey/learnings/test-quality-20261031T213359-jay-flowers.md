---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-10-31T21:33:59Z
identity: test-quality-20261031T213359-jay-flowers
tier: draft
---

Package-level claims that "warning prose is unpinned" must be verified per warning code, not per package. vibe-check PR #75 (ripwireadapter): a review comment stated the golden comparison reduced Warnings to a code-only list and that message text — error budgets, denylist listings, path names — could drift silently with only codes pinned. Verification at head showed the codes-only golden comparison was true (and is documented convention C5 in golden-notes.md, with placeholder messages in the goldens), but every cited dynamic segment was already pinned by per-warning substring assertions elsewhere in the package: budget_test.go ('1.9%', '2.2%', '0.16%'), denylist_test.go (pruned-directory listing), degrade_test.go (parse-error path 'pkg/bad.py'), guard_test.go (edge/source counts), cycles_test.go (corroboration phrases), tsabstractness_test.go ('abstract class', 'alias'). The true residue was four static disclosure strings (directory-as-module, name-resolved-edges, ambiguity-rate, unsupported-metric) with codes-only assertions. Tally rule for assessing such feedback: enumerate every warning code the package emits and grep each warning constructor's test file for a strings.Contains(w.Message, ...) assertion before accepting any "prose can drift silently" claim — prose assertions land only where a task's verification criterion named the message text, so coverage is uneven by construction, and a codes-only golden convention makes them the only prose pin.
