---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-10-27T16:21:27Z
identity: ci-cd-20261027T162127-jay-flowers
tier: draft
---

GitHub Actions reusable-workflow references must match the exact filename in the source repository — complytime/org-infra reusable workflows use the underscore `reusable_` prefix (reusable_release_preflight.yml, reusable_release_goreleaser.yml), consistent with convention CI-010. A prior learning (opsxgoreleaser-release-automation-20261004T141907-jay-flowers) incorrectly recorded hyphenated filenames; following it produced a release.yml whose dispatch failed with HTTP 422 \"workflow was not found\". When a stored learning contradicts the governing spec, re-verify against the live upstream repository at the pinned SHA before following either — the spec plus the live source are authoritative over recalled learnings, and stale learnings that contradict specs should be corrected at the source.
