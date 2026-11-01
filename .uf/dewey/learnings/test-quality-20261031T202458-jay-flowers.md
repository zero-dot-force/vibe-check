---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-10-31T20:24:58Z
identity: test-quality-20261031T202458-jay-flowers
tier: draft
---

When an adapter unconditionally emits warnings (e.g., error-budget and unsupported-metric in vibe-check's ripwireadapter), every graph it produces is status-partial by construction — so a Status==Partial assertion cannot detect the removal of any individual warning source. Regression tests for warning emission must assert the specific warning code (and message content), never the status alone. In the ripwireadapter zero-coupling-guard e2e test, the status assertion is transitively guaranteed by the ever-present unsupported-metric warning; only the warning-code assertion kills the mutation (removing the guardWarnings append).
