---
tag: hermetic-testing
author: jay-flowers
category: pattern
created_at: 2026-10-31T02:04:54Z
identity: hermetic-testing-20261031T020454-jay-flowers
tier: draft
---

When production code sanitizes the subprocess environment (e.g., metrics.SanitizeEnvironment strips everything but PATH/HOME/LANG), test control variables cannot reach a test double through cmd.Env. Bake them into the installed double instead: a PATH-installed wrapper script that sets the FAKE_* control variables and execs the committed fake binary. The control plane becomes a property of the installed double, not the parent process, so production code never names test-only env vars and the sanitized environment cannot strip them. Applied in vibe-check's internal/ripwireadapter fake-ripwire harness (installFakeRipwireWrapper).
