---
tag: scaffold-drift-test
author: jay-flowers
category: pattern
created_at: 2026-10-04T15:50:06Z
identity: scaffold-drift-test-20261004T155006-jay-flowers
tier: draft
---

The two-embed-FS pattern in vibe-check's scaffold package requires maintenance synchronization between two data structures: the categories slice (driving deployment in Run()) and the FS list in assetPaths() (driving drift detection). Both encode the same information — which embedded filesystem corresponds to which asset category — but are independently maintained. When a third asset category is added, both locations must be updated. The design doc acknowledges this as a maintenance note (D2), but the risk is latent in the code rather than structural. A categoriesWithFS struct pairing each category with its fs.FS would provide a single source of truth. This pattern mirrors the gaze project's approach but with two separate embed directives instead of one.
