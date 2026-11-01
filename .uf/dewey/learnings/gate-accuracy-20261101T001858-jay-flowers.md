---
tag: gate-accuracy
author: jay-flowers
category: gotcha
created_at: 2026-11-01T00:18:58Z
identity: gate-accuracy-20261101T001858-jay-flowers
tier: draft
---

When a cobra flag uses the explicit-opt-in pattern — Float64VarP(&v, name, "", defaultVal, help) plus `if cmd.Flags().Changed(name) { opts.X = &v }` — the registered cobra default is display-only: the effective default is "unset / no check" because the opts field stays nil until the user names the flag. vibe-check PR #75's README documented --max-duplication's Default column as 5.0 (copied from the registration site) while every sibling threshold row correctly said "unset (no check)" — a documented gate the code never enforces by default. Reviewer check: for any flag documented with a non-unset default, grep the registration site for a Changed() guard; the guard inverts the documented meaning.
