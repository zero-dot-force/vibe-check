---
tag: ref-validation
author: jay-flowers
category: gotcha
created_at: 2026-10-31T22:37:51Z
identity: ref-validation-20261031T223751-jay-flowers
tier: draft
---

Ref-topology gotcha for agents with restricted bash allowlists: `git merge-base --is-ancestor A B` communicates its answer ONLY through its exit code (0 = ancestor, 1 = not) and prints nothing either way — and the opencode bash tool renders a silent exit-1 identically to a silent exit-0 ("(no output)" success). This session that misread produced a false "PR already merged into main" conclusion until corrected. When ref ancestry matters, use commands whose answer is stdout, not exit status: `git merge-base A B` (compare the printed SHA against A) or `git rev-parse --verify`. If --is-ancestor must be used, pair it with an echo of $? in the same command so the code is visible in captured output.
