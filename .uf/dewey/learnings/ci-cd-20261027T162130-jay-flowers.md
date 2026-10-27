---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-10-27T16:21:30Z
identity: ci-cd-20261027T162130-jay-flowers
tier: draft
---

complytime/org-infra reusable release workflows follow the CI-010 underscore convention — reusable_release_preflight.yml and reusable_release_goreleaser.yml (verified at pinned SHA 58194ad2d1b24803263cac6fd1b54422e3c208dd). Hyphenated variants do not exist upstream: they pass commit-time YAML validation and surface only at dispatch as HTTP 422 \"workflow was not found\", so local YAML linting cannot catch the mismatch. This supersedes the 2026-10-04 draft learning opsxgoreleaser-release-automation-20261004T141907-jay-flowers, which claimed the opposite and likely caused the original defect. Two guardrails when wiring cross-repo reusable workflows: (1) resolve the exact workflow filename at the pinned SHA before merging; (2) when a draft-tier Dewey learning contradicts the governing spec, the spec prevails — reconcile the learning, not the code. Also: extraneous with: inputs (e.g., go-version-file on the preflight workflow) are not declared by the upstream reusable workflow and must not be passed.
