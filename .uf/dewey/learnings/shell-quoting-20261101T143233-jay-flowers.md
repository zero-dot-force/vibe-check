---
tag: shell-quoting
author: jay-flowers
category: gotcha
created_at: 2026-11-01T14:32:33Z
identity: shell-quoting-20261101T143233-jay-flowers
tier: draft
---

fmt %q is not shell-safe for embedding filesystem paths in POSIX sh wrapper scripts: %q renders double quotes, and sh expands $variable and backtick substitutions inside double quotes, so an adversarial checkout path or TMPDIR value can execute arbitrary commands when the wrapper runs. Use POSIX single-quote escaping instead (wrap the whole word in single quotes, replace every embedded single quote with the '\'' sequence), ideally via a named helper (shellSingleQuote) so the pattern is pinned and reviewable. _test.go helpers are not importable across packages, so duplicate the helper per test package and promote it to a shared test-support package only at the third consumer. Encoded in vibe-check PR #75 (cmd/vibe-check/cli_ripwire_test.go wrapper hardening, replacing fmt %q wrappers).
