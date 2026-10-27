---
tag: opsxgoreleaser-release-automation
author: jay-flowers
created_at: 2026-10-04T14:19:07Z
identity: opsxgoreleaser-release-automation-20261004T141907-jay-flowers
tier: draft
---

CORRECTED 2026-10-27: the org-infra reusable workflow files use UNDERSCORES (reusable_release_preflight.yml, reusable_release_goreleaser.yml), verified live at pinned SHA 58194ad2d1b24803263cac6fd1b54422e3c208dd — NOT hyphens, as this learning originally and wrongly claimed. The original hyphenated claim was recorded from an assumed naming convention without checking the live repository; a release.yml built on it failed dispatch with HTTP 422 "workflow was not found". Always resolve the actual filenames in the upstream org-infra repository at the pinned SHA rather than assuming a convention, and when a stored learning contradicts the governing spec (openspec/changes/goreleaser-release-automation/), the spec plus the live source prevail. Superseding learnings: github-actions-20261027T162125-jay-flowers, ci-cd-20261027T162127-jay-flowers, ci-cd-20261027T162130-jay-flowers, ci-cd-20261027T162132-jay-flowers.
