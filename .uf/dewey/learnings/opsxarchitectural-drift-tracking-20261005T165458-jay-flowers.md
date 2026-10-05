---
tag: opsxarchitectural-drift-tracking
author: jay-flowers
category: gotcha
created_at: 2026-10-05T16:54:58Z
identity: opsxarchitectural-drift-tracking-20261005T165458-jay-flowers
tier: draft
---

When using os/exec in Go to suppress command output, setting cmd.Stderr = nil does NOT discard stderr — Go's os/exec connects nil streams to the parent process's stderr by default. To truly discard stderr, use cmd.Stderr = io.Discard. This was caught during code review on the git metadata resolution helper (runGitCmd in cmd/vibe-check/analyze.go). The comment said "Discard stderr to avoid noise from git when not in a repo" but nil doesn't achieve that — io.Discard does. This is a subtle Go os/exec API behavior: nil = inherit parent, io.Discard = suppress.
