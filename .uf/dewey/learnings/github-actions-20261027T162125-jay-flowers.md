---
tag: github-actions
author: jay-flowers
category: gotcha
created_at: 2026-10-27T16:21:25Z
identity: github-actions-20261027T162125-jay-flowers
tier: draft
---

complytime/org-infra reusable release workflows use underscore filenames (reusable_release_preflight.yml, reusable_release_goreleaser.yml) as of pinned SHA 58194ad2d1b24803263cac6fd1b54422e3c208dd. A prior draft-tier learning (opsxgoreleaser-release-automation-20261004T141907-jay-flowers) asserts the opposite — hyphenated names, "NOT underscores" — and was never verified against the live repository; release.yml implemented per that learning failed dispatch with HTTP 422 "workflow was not found". That learning is superseded by this one. Root gotcha: GitHub validates neither reusable-workflow filenames nor with: inputs against the target repository at commit or lint time — both defects surface only at dispatch/run time (the 422, and the undeclared go-version-file input that the preflight workflow never accepted). When writing or reviewing a job-level uses: reference to a reusable workflow: (1) resolve the filename against the target repository at the pinned SHA; (2) check every with: input against the called workflow's declared workflow_call inputs (preflight accepts only tag — Go setup from go.mod happens inside reusable_release_goreleaser.yml per the release-workflow spec); (3) prefer authored spec/design artifacts and the CI-010 reusable_ naming convention over unverified draft-tier learnings, and record filename claims as learnings only after live confirmation at the pinned ref.
