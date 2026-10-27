---
tag: release-automation
author: jay-flowers
category: gotcha
created_at: 2026-10-27T19:42:07Z
identity: release-automation-20261027T194207-jay-flowers
tier: draft
---

complytime/org-infra reusable_release_preflight.yml (pinned SHA 58194ad2d1b24803263cac6fd1b54422e3c208dd) declares TWO relevant inputs: tag and ci_checks — superseding the earlier "preflight accepts only tag" claim. Its CI-check auto-discovery recognizes ONLY org-infra's own workflow filename conventions (ci_checks.yml, ci_local.yml, ci_security.yml); a repo whose CI lives in differently-named workflows (e.g. ci.yml) gets an empty required-checks list, which the preflight materializes as [""] and aborts with `Check '' not found` at "Verify CI checks passed on HEAD" — before any tag is created (fails closed but blocks all releases). Fix: pass ci_checks explicitly as a single-quoted JSON array of BARE job check-run names (no workflow-name prefix, no matrix suffix) exactly matching check-run names on the HEAD commit. Check-run names equal bare job ids even when the workflow has a top-level name (verified empirically on zero-dot-force/vibe-check main HEAD 34bca97: names are `test`, `lint`). pull_request-only workflows must be excluded (no check-runs on main HEAD), and ci.yml must run on push to main. Safe pre-merge probe: dispatch the release workflow from the feature branch with a scratch semver tag — verification fails on the branch HEAD before tag creation, validating input parsing with zero side effects (vibe-check probe run 40568510866: REQUIRED_CHECKS echoed ["lint","test"]).
