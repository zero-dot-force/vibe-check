---
tag: release-automation
author: jay-flowers
category: gotcha
created_at: 2026-10-27T19:40:43Z
identity: release-automation-20261027T194043-jay-flowers
tier: draft
---

complytime/org-infra reusable_release_preflight.yml (pinned SHA 58194ad2d1b24803263cac6fd1b54422e3c208dd) declares a ci_checks input alongside tag. Its CI-check auto-discovery recognizes ONLY org-infra's own workflow filename conventions (ci_checks.yml, ci_local.yml, ci_security.yml); a repo whose CI lives in differently-named workflows (vibe-check: ci.yml) gets an empty required-checks list, which materializes as [""] via printf|jq, and "Verify CI checks passed on HEAD" aborts with `Check '' not found` — BEFORE any tag is created (fails closed but blocks all releases). Fix: pass ci_checks explicitly as a single-quoted JSON array of bare job check-run names — no workflow-name prefix (the "CI / test" UI prefix is not part of the check-run name), no matrix " (combo)" suffixes — exactly matching check-runs on the HEAD commit; derive the list from push-triggered workflows only (pull_request-only gates like structural-gate never produce check-runs on a main HEAD and would fail every dispatch); the CI workflow must run on push to main so check-runs exist at dispatch time. vibe-check: ci_checks: '["lint", "test"]' matching ci.yml job IDs (build+vet are steps inside test). Safe pre-merge probe pattern for reusable-workflow input changes: dispatch the release workflow from the feature branch with a scratch semver tag (e.g. v0.1.99) — CI verification fails on the branch HEAD before tag creation with zero side effects, proving input acceptance and REQUIRED_CHECKS parsing (probe run 40568510866: REQUIRED_CHECKS ["lint","test"]). This supersedes the earlier "preflight accepts only tag" claim — declared inputs at that SHA are tag + ci_checks (+ optional allow_prerelease/skip_* flags).
