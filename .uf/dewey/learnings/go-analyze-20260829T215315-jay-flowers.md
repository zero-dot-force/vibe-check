---
tag: go-analyze
author: jay-flowers
category: pattern
created_at: 2026-08-29T21:53:15Z
identity: go-analyze-20260829T215315-jay-flowers
tier: draft
---

The exitCodeError pattern is the correct way to handle process exit codes in cobra CLI applications without calling os.Exit directly in RunE handlers. Define a type `exitCodeError struct { code int; err error }` implementing Error() and Unwrap(), return it from RunE, and extract it with errors.As in main(). This preserves deferred cleanup (signal handlers, context cancellation), makes the full cobra execution path testable, and keeps os.Exit() isolated to main(). The RunAnalyze function (AP-002 testable entry point) returns the exit code in the result struct, while the cobra layer wraps it in exitCodeError for main() to process.
