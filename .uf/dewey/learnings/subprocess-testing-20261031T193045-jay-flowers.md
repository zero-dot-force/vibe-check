---
tag: subprocess-testing
author: jay-flowers
category: gotcha
created_at: 2026-10-31T19:30:45Z
identity: subprocess-testing-20261031T193045-jay-flowers
tier: draft
---

Go's %q verb is not a shell-quoting function. When generating #!/bin/sh wrapper scripts by string interpolation, %q produces double-quoted strings in which POSIX sh expands $ and backticks — a checkout path or TMPDIR containing those characters silently changes the script's meaning. Use a POSIX single-quote escaper (replace ' with '\'' and wrap in single quotes) instead, as vibe-check's internal/ripwireadapter/fake_ripwire_test.go shellSingleQuote does; the CLI-level copy in cmd/vibe-check/cli_ripwire_test.go that used %q is the counter-example. Same class: os.WriteFile(path, data, 0o755) is masked by the process umask, so test-installed executables need an explicit os.Chmod to pin the executable bit.
