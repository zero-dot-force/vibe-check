---
tag: subprocess-security
author: jay-flowers
category: gotcha
created_at: 2026-11-01T00:18:56Z
identity: subprocess-security-20261101T001856-jay-flowers
tier: draft
---

Go's %q renders double-quoted strings, and POSIX sh expands $, backticks, and backslashes inside double quotes — so fmt.Sprintf("VAR=%q exec %q \"$@\"", dir, bin) in a generated #!/bin/sh script is a command-injection vector whenever an interpolated path carries adversarial characters (a hostile TMPDIR or checkout path containing $(...) or backticks). vibe-check PR #75 hit this twice in one test harness (the fake-riprewire PATH wrapper and the oversized-output flood script). The safe idiom is POSIX single-quoting — wrap in '…' and replace embedded single quotes with '\''' — as a shellSingleQuote(s) helper. Pair it with an explicit os.Chmod(script, 0o755) after os.WriteFile: a restrictive umask strips the exec bit, exec.LookPath then silently skips the script, and negative-path assertions pass vacuously.
