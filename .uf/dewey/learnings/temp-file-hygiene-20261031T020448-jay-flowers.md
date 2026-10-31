---
tag: temp-file-hygiene
author: jay-flowers
category: gotcha
created_at: 2026-10-31T02:04:48Z
identity: temp-file-hygiene-20261031T020448-jay-flowers
tier: draft
---

os.Remove on a file never removes its os.MkdirTemp parent directory — the Go standard library does not garbage-collect empty temp dirs. The recurring failure mode (found in vibe-check internal/ripwireadapter/assembly.go writeMinimalArchRules, 2026-10-30): create MkdirTemp dir, write file inside, defer os.Remove(filePath), and write a doc comment claiming "the directory is removed with it by the os package once empty" — false; every run leaks one temp dir. Correct pattern: capture the dir, defer os.RemoveAll(dir) (covers both file and dir), and delete any comment asserting os auto-cleanup. Audit new code for MkdirTemp call sites whenever only a file path is deferred for cleanup.
