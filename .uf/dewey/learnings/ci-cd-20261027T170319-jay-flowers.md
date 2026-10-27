---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-10-27T17:03:19Z
identity: ci-cd-20261027T170319-jay-flowers
tier: draft
---

A reusable-workflow reference fix in a release pipeline is proven only by an end-to-end workflow_dispatch: GitHub validates neither the `uses:` filename nor unexpected `with:` inputs at commit or PR-CI time — both surface only at dispatch (HTTP 422 "workflow was not found" or an unexpected-input failure) — so inspection-verified corrections remain unproven until a run actually resolves the reusable workflows. A side-effect-free verification exists: dispatch with a deliberately invalid tag (e.g. "verify-dispatch"). Workflow resolution succeeds (proving the reference fix), and the preflight's semver validation fails before any tag is created or any release artifact is built. Execute this against the fix branch (`gh workflow run <workflow> --ref <branch>`, valid while the workflow file is registered on the default branch) or immediately post-merge, before relying on the pipeline for a real release.
