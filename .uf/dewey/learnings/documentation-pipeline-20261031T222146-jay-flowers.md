---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-10-31T22:21:46Z
identity: documentation-pipeline-20261031T222146-jay-flowers
tier: draft
---

When fixing stale CLI help text for a capability backed by fail-closed pinned external-tool constants (vibe-check PR #75 Item 13: ripwire 0.6.9@bedc0cd), do not hardcode the pin pair into cobra Short/Long strings. The pin's single source of truth is the adapter package's unexported constants (internal/ripwireadapter/version.go: pinnedRipwireVersion/pinnedRipwireCommit), which the absent-binary error interpolates and cli_ripwire_test.go asserts — an unguarded literal copy in help text goes silently stale on the next pin bump, recreating the exact stale-help finding class the fix was meant to close. Either phrase the help text generically ("requires a pinned, validated ripwire build; the error output and README name the exact version") or export an accessor from the adapter package and interpolate at command construction; keep the version literal out of Short entirely (cobra Short is a one-line summary, and the feedback's intent — discoverability of the requirement — is met by Long + pointer to the authoritative surfaces).
