---
tag: add-divisor-entropy-agent
author: jay-flowers
category: pattern
created_at: 2026-09-01T02:34:23Z
identity: add-divisor-entropy-agent-20260901T023423-jay-flowers
tier: draft
---

On 2026-08-31 in the vibe-check repo (branch add-divisor-entropy-agent), the three CLI subcommands analyze, diff, and init were all built on the same testable architecture pattern (category: pattern): a RunXxx(ctx context.Context, opts XxxOptions) (*XxxResult, error) function holds ALL logic — flag validation, calling the domain layer, exit-code mapping (0 success, 1 policy violation for analyze, 2 tool/IO failure), and a single buffered stdout write guarded by ctx.Err() so no partial output is emitted on cancellation — while the cobra RunE is only a thin wrapper that wires cmd.OutOrStdout()/cmd.ErrOrStderr(), binds flags, sets up signal.NotifyContext(SIGINT/SIGTERM), and returns an exitCodeError{code,err} consumed by main() for the process exit code. Crucially, an unexported injectable seam — writeFile func(path string, data []byte, perm fs.FileMode) error on the options struct, defaulting to os.WriteFile when nil — lets tests deterministically exercise the I/O-failure exit-2 branch without needing root or a read-only filesystem. init reuses diff's writeListSection helper (DRY), and embed.go mirrors metrics/schema.go's //go:embed var + exported accessor so staticcheck's unused rule stays green.
