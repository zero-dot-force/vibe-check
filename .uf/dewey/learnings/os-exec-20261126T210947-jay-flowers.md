---
tag: os-exec
author: jay-flowers
category: pattern
created_at: 2026-11-26T21:09:47Z
identity: os-exec-20261126T210947-jay-flowers
tier: draft
as_of: "vibe-check@40a2e0eb4a5313ddb15330ed799f48435eca3ad7"
---

A Go subprocess runner that drains cmd.StdoutPipe/StderrPipe to EOF before cmd.Wait deadlocks whenever a grandchild inherits the pipes: on oversize or context cancellation only the direct child is killed, EOF never arrives, and the run blocks forever (golang/go#23019). The standard-library fix is cmd.WaitDelay plus non-*os.File writers on cmd.Stdout/cmd.Stderr: os/exec closes abandoned pipes once the delay elapses after process exit or context done, and a writer that returns an error on crossing a capture bound makes os/exec close that stream immediately, EPIPE-ing a still-writing grandchild. Classify a zero-exit run that surfaces exec.ErrWaitDelay as a hard error (the capture may be incomplete — fail closed), and keep error precedence: context cancellation first, then oversize, then ErrWaitDelay, then exit status. Guard every construction path: build runners through a defaulted constructor, and fall back to the default delay at the use site when the field is zero so struct literals cannot silently lose the bound.
