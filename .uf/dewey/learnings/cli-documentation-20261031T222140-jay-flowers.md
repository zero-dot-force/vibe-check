---
tag: cli-documentation
author: jay-flowers
category: gotcha
created_at: 2026-10-31T22:21:40Z
identity: cli-documentation-20261031T222140-jay-flowers
tier: draft
---

When user-facing CLI help must state an external-tool version pin, derive it from the enforcement constants instead of duplicating the literal. In vibe-check PR #75's feedback fix (cobra help text for the now-multi-language `analyze` command), embedding "ripwire 0.6.9@bedc0cd" in analyzeCmd Long would duplicate internal/ripwireadapter/version.go's unexported pinnedRipwireVersion/pinnedRipwireCommit — a second source of truth that drifts silently on the next pin bump, and the unexported constants give the CLI layer no interpolation path. Options: export the pinned pair from the internal package and Sprintf it into the Long, or phrase the help generically and leave the literal to code, error guidance, and README. Corollary: cobra help tests typically assert command and flag names rather than Short/Long prose (verified in vibe-check: the task-7.5 help tests and PR #75's new CLI tests assert flags, exit codes, and guidance strings), so help-string fixes are behavior-neutral — but grep for the replaced prose before editing to confirm.
