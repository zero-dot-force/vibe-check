---
tag: test-coverage
author: jay-flowers
category: gotcha
created_at: 2026-10-29T15:54:21Z
identity: test-coverage-20261029T155421-jay-flowers
tier: draft
---

When a fix round adds defensive early-return guards to parallel code paths, enumerate the guards in the changed code and diff them against the added tests before closing the round. In vibe-check's init UF-registration change (code review iteration 2), spliceCapabilities and spliceMatrix each gained no-section and inline-value guards, but only the capabilities side received fixtures; spliceMatrix's no-advisor-section and inline-advisor guards shipped untested even though the spec scenario "Absent or shape-failed units are notice-only" names "the review matrix lacks the advisor: section". Fix-round test additions tend to mirror the fix walkthrough's examples rather than the full guard surface of every touched path.
