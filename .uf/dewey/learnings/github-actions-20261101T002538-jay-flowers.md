---
tag: github-actions
author: jay-flowers
category: gotcha
created_at: 2026-11-01T00:25:38Z
identity: github-actions-20261101T002538-jay-flowers
tier: draft
---

Static GitHub Actions templates that hardcode `on: pull_request: branches: [main]` silently never run in repositories whose default branch is not main — no error, no notice, no run history. If the consumer marks the check required in branch protection (as onboarding guidance often instructs), PRs deadlock waiting for a check that never reports. This is a second member of the silent-gate-failure family alongside the vacuous-green-gate (non-PR trigger + empty base context): both fail without any signal. Templates meant for arbitrary repos should either omit the branches filter (pull_request-only still guarantees github.event.pull_request.base.sha context) or detect and notice the default-branch assumption (e.g. git symbolic-ref refs/remotes/origin/HEAD), mirroring vibe-check init's manifest-detection notice pattern. Spec scenarios should pin artifact bytes ('contains only pull_request with branches: [main]') rather than unobservable intent ('targeting the default branch').
