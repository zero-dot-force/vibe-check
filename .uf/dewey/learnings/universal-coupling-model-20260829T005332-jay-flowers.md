---
tag: universal-coupling-model
author: jay-flowers
category: gotcha
created_at: 2026-08-29T00:53:32Z
identity: universal-coupling-model-20260829T005332-jay-flowers
tier: draft
---

When testing Go subprocess communication (ExternalAdapter in vibe-check), the TestHelperProcess pattern (os.Args[0] with GO_WANT_HELPER_PROCESS env var) is the standard library-compatible approach that avoids external test binaries. Key gotchas: (1) the test function parameter should use _ *testing.T since it's not a real test, (2) tests using t.Setenv cannot use t.Parallel — Go 1.17+ enforces this, (3) limitedBuffer for stderr capture must always report the full write length to avoid breaking the subprocess pipe even on internal errors, (4) Gaze quality analysis revealed that constructor functions (NewExternalAdapter, NewRegistry) often have 0% contract coverage because tests call them but never directly assert on the return value — adding nil checks and default value assertions closes this gap easily.
