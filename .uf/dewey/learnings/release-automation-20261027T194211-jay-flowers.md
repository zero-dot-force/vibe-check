---
tag: release-automation
author: jay-flowers
category: pattern
created_at: 2026-10-27T19:42:11Z
identity: release-automation-20261027T194211-jay-flowers
tier: draft
---

An explicit ci_checks allowlist in a release preflight pins the release gate to the CI job set at commit time, and the two drift modes behave asymmetrically: renaming/removing a job fails loudly at preflight (missing check name aborts before tag creation), but ADDING a new CI job is silent — the new job does not gate releases until the allowlist is updated, and nothing signals the gap. Document the sync obligation at BOTH ends (comment beside the ci_checks input in release.yml AND a marker comment in the CI workflow near the job IDs, e.g. "Job IDs feed release.yml preflight ci_checks verbatim — update ci_checks when renaming or adding jobs"), covering additions explicitly. Prefer an automated drift check asserting the allowlist equals the CI workflow's job-key set so additions surface in CI instead of at the next release dispatch. Also record input changes in tasks.md (task-level record) alongside spec/design/checklist/CHANGELOG when remediating requirement-chain artifacts.
