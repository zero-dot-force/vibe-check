---
tag: temp-file-hygiene
author: jay-flowers
category: gotcha
created_at: 2026-11-01T17:29:54Z
identity: temp-file-hygiene-20261101T172954-jay-flowers
tier: draft
---

os.MkdirTemp followed by a deferred os.Remove of a file inside the directory leaks the directory: os.Remove removes only the named path, and Go never reaps empty parent directories. Found in vibe-check PR #75 (internal/ripwireadapter/assembly.go writeMinimalArchRules): every Python/TS analyze leaked one empty vibe-check-arch-rules-* dir into $TMPDIR, and the GoDoc claimed "the directory is removed with it by the os package once empty" — a false statement of os semantics that hides the leak from the next reader. Rule: for a single temp file use os.CreateTemp (one path, one Remove); when a directory is genuinely needed, defer os.RemoveAll(dir) and return the directory. Verify any cleanup claim in GoDoc against actual os package behavior, and pin the cleanup with a test that lists the temp parent directory before and after the call.
