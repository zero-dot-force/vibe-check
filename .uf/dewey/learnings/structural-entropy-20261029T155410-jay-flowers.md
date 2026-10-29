---
tag: structural-entropy
author: jay-flowers
category: pattern
created_at: 2026-10-29T15:54:10Z
identity: structural-entropy-20261029T155410-jay-flowers
tier: draft
---

Verified on vibe-check init-registers-divisor-entropy (base 3e354b7 → head c9fb41b): adding same-package files with no new imports and no new exported or method-bearing types produces a zero base→PR structural delta in vibe-check's Martin metrics. internal/scaffold grew 206→740 totalLines across three new files yet Ca/Ce/Instability/Abstractness/Distance/LCOM held exactly (Ce 9, I 0.9, LCOM 0). Recipe for a green entropy dogfood gate on additive growth: (1) place new code in an existing package rather than a new module; (2) restrict imports to packages the package already has so efferent coupling cannot move; (3) keep new types unexported and method-less (data structs + standalone functions) so exportedTypes, abstractness, and LCOM4 stay flat; (4) in-package file splitting is structurally free. Caveat for reviewers: a zero delta means structure held, not that nothing changed — totalLines and the diff prove the growth is real.
