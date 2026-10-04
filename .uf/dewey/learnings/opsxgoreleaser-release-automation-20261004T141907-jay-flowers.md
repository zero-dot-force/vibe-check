---
tag: opsxgoreleaser-release-automation
author: jay-flowers
created_at: 2026-10-04T14:19:07Z
identity: opsxgoreleaser-release-automation-20261004T141907-jay-flowers
tier: draft
---

When adopting patterns from unbound-force/gaze, the org-infra reusable workflow files use hyphens (reusable-release-preflight.yml, reusable-release-goreleaser.yml) and NOT underscores. Always verify the actual filenames in the upstream org-infra repository rather than assuming a naming convention. The spec and implementation artifacts should be cross-checked against the live source files to catch filename mismatches before code review.
