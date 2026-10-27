---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-10-27T16:21:32Z
identity: ci-cd-20261027T162132-jay-flowers
tier: draft
---

The complytime/org-infra reusable release workflows use underscore filenames — reusable_release_preflight.yml and reusable_release_goreleaser.yml — verified at pinned commit 58194ad2d1b24803263cac6fd1b54422e3c208dd (October 2026). This supersedes the 2026-10-04 learning claiming hyphenated filenames (reusable-release-preflight.yml / reusable-release-goreleaser.yml): those references failed dispatch with HTTP 422 'workflow was not found'. GitHub Actions resolves `uses:` references to external reusable workflows only at run time — never at commit time, and never in PR CI for a workflow_dispatch-only file like release.yml — and unknown `with:` inputs (e.g., go-version-file passed to the preflight workflow) likewise fail only at dispatch. Before pinning or bumping a reusable-workflow SHA, resolve the workflow filenames and the declared input list against the target repository at that exact commit, and re-verify on every pin update.
