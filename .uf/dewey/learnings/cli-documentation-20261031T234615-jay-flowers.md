---
tag: cli-documentation
author: jay-flowers
category: gotcha
created_at: 2026-10-31T23:46:15Z
identity: cli-documentation-20261031T234615-jay-flowers
tier: draft
---

When porting prose from spec/design documents into user-facing CLI help text, grep the canonical spelling of every proper noun (tool names, flag names, URLs) against the code's ground-truth surfaces before shipping: the adapter package path, the exec.LookPath target, and the tool's own recorded --version output. Spec docs can carry latent misspellings that help text then inherits — in vibe-check PR #75's item-13 help-text fix, the analyze Long text spelled the external binary "riprewire" (rip+rewire) on two lines while the Short description, the internal/ripwireadapter package, the runner's LookPath, the recorded version fixtures, and README all use "ripwire" (rip+wire); the deviant spelling traced back to misspellings in the ripwire-adapter change's own design.md and tasks.md. The binary's self-reported version output is the naming ground truth, and because help tests assert flags and guidance strings rather than Long prose, a repo-wide grep for the deviant form (which matched only the two help-text lines among Go sources) confirms the fix is surgery-free.
