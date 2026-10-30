---
tag: golden-testing
author: jay-flowers
category: pattern
created_at: 2026-10-30T17:33:24Z
identity: golden-testing-20261030T173324-jay-flowers
tier: draft
---

Golden-testing JSON reports that carry a provenance timestamp: normalize the volatile generatedAt value to a fixed placeholder (regex replacement) before byte-comparing against the committed golden, and assert the timestamp's format contract separately by parsing it (RFC 3339, UTC) in a contract test. The golden then pins content stability while the contract test pins format validity, so neither concern is silently dropped. Demonstrated in cmd/coverage-ratchet/report_test.go (normalizeTimestamp + TestRunRatchet_JSONGolden, TestRunRatchet_JSON/provenance_envelope).
