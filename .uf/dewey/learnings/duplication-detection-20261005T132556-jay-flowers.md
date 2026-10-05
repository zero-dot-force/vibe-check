---
tag: duplication-detection
author: jay-flowers
created_at: 2026-10-05T13:25:56Z
identity: duplication-detection-20261005T132556-jay-flowers
tier: draft
---

When normalizing Go ASTs for structural comparison (duplication detection), in-place AST mutation (replacing *ast.Ident.Name and *ast.BasicLit.Value before formatting) is the correct approach. The alternative — formatting first then doing string replacement on the formatted output — causes corruption when identifier names appear as substrings inside string literals. However, in-place mutation creates an ordering dependency: any code running after normalization will see corrupted identifiers. The GoDoc should clearly document this destructive behavior, and the pipeline order must ensure normalization runs after all other AST consumers. For vibe-check's goadapter, the pipeline order is: type classification, LCOM, extensions, then duplication detection, which is safe.
