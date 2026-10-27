---
tag: release-automation
author: jay-flowers
category: pattern
created_at: 2026-10-27T19:40:46Z
identity: release-automation-20261027T194046-jay-flowers
tier: draft
---

An explicit ci_checks allowlist in a release preflight pins the release gate to the CI job set at commit time, and the two drift modes behave asymmetrically: renaming/removing a job fails loudly at preflight (missing check name aborts before tag creation), but ADDING a new CI job is silent — the new job does not gate releases until the allowlist is extended, with no signal of the gap. When documenting the sync obligation, cover additions explicitly ("if job names change or jobs are added"). Record the coupling at BOTH ends: a comment beside the ci_checks input (consumer side) and a marker comment in the CI workflow above the jobs whose IDs are load-bearing (mutation side). Longer term, prefer a drift check that asserts the allowlist equals the CI workflow's push-triggered job-key set so additions surface in CI instead of at the next release dispatch.
