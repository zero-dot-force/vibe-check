---
tag: github-actions
author: jay-flowers
category: gotcha
created_at: 2026-10-27T17:52:46Z
identity: github-actions-20261027T175246-jay-flowers
tier: draft
---

GitHub Actions reusable-workflow permissions cut both ways: a called workflow cannot elevate beyond the caller's grant, but it runs with (inherits) whatever the caller's job grants — the callee's own permissions declaration can only narrow further. Two consequences: (1) a caller whose only permissions block is a workflow-level `permissions: {}` (default-deny) makes every reusable-workflow job fail at startup before any step runs (run reports startup_failure with no job logs); (2) over-granting at the caller silently widens the callee's effective GITHUB_TOKEN scopes. Least-privilege pattern: keep the workflow-level `permissions: {}` as default-deny and grant each reusable-workflow job exactly the scopes the callee declares at its pinned SHA — no broader — with an inline comment anchoring the mirror to the pinned SHA so any pin bump forces re-verification of the granted scopes. Like `uses:` filenames and `with:` inputs, permission mismatches are validated by GitHub only at dispatch/run time, never at commit or PR-CI time; a deliberately invalid-tag dispatch (fails safely at preflight tag validation, creates nothing) proves workflow resolution and job startup without side effects. Verified in zero-dot-force/vibe-check: run 40547660936 startup_failure with empty top-level permissions → run 40548509351 jobs start after job-level grants (preflight contents:write + checks:read; goreleaser release contents:write + id-token:write for cosign/syft).
