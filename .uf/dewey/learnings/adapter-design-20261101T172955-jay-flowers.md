---
tag: adapter-design
author: jay-flowers
category: pattern
created_at: 2026-11-01T17:29:55Z
identity: adapter-design-20261101T172955-jay-flowers
tier: draft
---

When an analysis pipeline composes per-task scan functions over the same corpus, each task implementing its own filepath.WalkDir + os.ReadFile multiplies corpus I/O linearly with task count. In vibe-check PR #75's ripwire adapter, the Python path read every .py file twice (pythonAbstractnessScan and lcom4Scan each ReadFile the same files) and walked the crawl root up to five times (those two scans plus countSourceFiles, detectDenylistPruned, and crawl-root resolution) — on top of three ripwire subprocess crawls. Consolidate at the assembly seam: one walk that reads each file once and fans the bytes out to every extractor, with census-style passes folded into the same walk. Preserve documented per-scan semantic differences (e.g., a safety-net census that deliberately ignores the denylist so it over-counts in the safe direction) by parameterizing the shared walk rather than duplicating it.
