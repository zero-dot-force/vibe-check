---
tag: vibe-check
author: jay-flowers
category: gotcha
created_at: 2026-10-02T19:58:21Z
identity: vibe-check-20261002T195821-jay-flowers
tier: draft
---

vibe-check's `analyze` command takes a directory path (default `.`), NOT a go/packages `./...` glob pattern. The goadapter's resolvePackages sets `cfg.Dir = projectPath` and then hardcodes `packages.Load(cfg, "./...")`, so the positional argument is the working directory, not the package pattern. Moreover `metrics.ValidateProjectPath` calls EvalSymlinks+Stat on the literal path and rejects any non-directory string like `./...` with "path does not exist". So a CI workflow must invoke `vibe-check analyze --output head.json .` (and `./base` for a base checkout), never `./...`. Note: AGENTS.md and README examples still show `./...` in some places — this is a pre-existing doc inaccuracy, not the correct invocation. Discovered during the ci-regression-gate change dry-run.
