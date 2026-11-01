---
tag: test-infrastructure
author: jay-flowers
category: gotcha
created_at: 2026-10-31T19:31:00Z
identity: test-infrastructure-20261031T193100-jay-flowers
tier: draft
---

Go's fmt %q emits Go-syntax double-quoted strings, which are NOT safe to interpolate into POSIX shell scripts: inside sh double quotes, '$' triggers variable expansion and '`' triggers command substitution, and %q escapes neither. A test wrapper generated with Sprintf("...FAKE_RIPWIRE_RECORDINGS=%q...", path) breaks — or becomes an injection surface — when the repo path contains those characters. Use a single-quote escaper (replace ' with '\'' and wrap in single quotes); internal/ripwireadapter's fake-ripwire harness shellSingleQuote is the model, while the cmd/vibe-check CLI copy of the same wrapper that used %q silently diverged as the fragile variant (found in PR #75 review, 2026-10-31).
