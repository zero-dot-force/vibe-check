---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-10-31T23:46:25Z
identity: test-quality-20261031T234625-jay-flowers
tier: draft
---

When a fix's entire value is a deterministic property of an error message — e.g. vibe-check ripwireadapter's requireInts refactor (PR #75 feedback commit e986773) from map iteration to ordered intAttr pairs, which makes the "missing X attribute" error name the first-declared attribute when several are absent — an `err != nil` table case over the same input verifies only the fail-closed behavior, not the determinism the refactor bought. Recipe for any ordering/determinism fix: add the assertion that fails under the pre-fix implementation. For map→slice error-ordering fixes, feed an input with ALL attributes missing at once and assert the first-declared name appears in the error (map iteration fails that assertion with probability 1−1/n per run). Symptom of the gap: the changed function's GoDoc promises determinism while the owning test file's malformed-input table asserts nothing about error text.
