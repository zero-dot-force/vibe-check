#!/usr/bin/env python3
"""Measure ripwire's TypeScript import-resolution rate on a crawl root.

openspec/changes/bump-ripwire-pin design D6 / tasks 5.1-5.2. Spec-phase
evidence, never built or run by CI. Run with `python3 -I`.

    python3 -I measure_typescript.py <ripwire-binary> <crawl-root> [--json]

<crawl-root> is the directory holding tsconfig.json / package.json, as the
adapter would choose it.

Measured: ripwire's own unresolved-import disclosure on --deps — the root
attribute imports_unresolved= (TS/JS imports that name the tree yet drew no
edge) — over the in-tree TS/JS import count. The Phase 0 spike measured
rollup @ ad39ddc on 0.6.9 at 14 / 8,529 = 0.16% unresolved. ripwire's TS
file graph carries no provenance weighting, so unresolved targets are the
only error class it discloses; external package imports are not errors and
are not counted.

The denominator is the number of <inc t=> rows on dependency-capable TS/JS
files whose target is relative or tsconfig-aliased (starts with "." or "/",
or does not look like a bare package specifier), counted over the full
--deps window (--limit is raised to cover the tree).
"""

import json
import subprocess
import sys
import xml.etree.ElementTree as ET

TS_EXTENSIONS = (".ts", ".tsx", ".js", ".jsx", ".mts", ".cts", ".mjs", ".cjs")


def is_in_tree_specifier(target):
    """Heuristic: relative or absolute paths name the tree; bare names are packages."""
    return target.startswith((".", "/"))


def main():
    if len(sys.argv) < 3:
        sys.exit(__doc__)
    ripwire, root = sys.argv[1], sys.argv[2]
    as_json = "--json" in sys.argv[3:]

    version = subprocess.run([ripwire, "--version"], capture_output=True, text=True, check=True).stdout.strip()
    xml = subprocess.run(
        [ripwire, ".", "--deps", "--limit=100000", "--no-cache"],
        cwd=root, capture_output=True, check=True,
    ).stdout.decode()
    deps = ET.fromstring(xml)
    if deps.get("has_more") not in (None, "0"):
        sys.exit("deps window did not cover the tree; raise --limit")

    in_tree = 0
    total = 0
    for f in deps.iter("f"):
        path = f.get("p", "")
        if not path.endswith(TS_EXTENSIONS):
            continue
        for inc in f.findall("inc"):
            total += 1
            if is_in_tree_specifier(inc.get("t", "")):
                in_tree += 1
    unresolved = int(deps.get("imports_unresolved", "0"))

    result = {
        "ripwire": version,
        "crawl_root": root,
        "files": int(deps.get("files", "0")),
        "inc_rows": total,
        "in_tree_inc_rows": in_tree,
        "imports_unresolved": unresolved,
        "graph_partial": deps.get("graph_partial") == "1",
        "unresolved_pct": round(100 * unresolved / max(1, in_tree), 2),
    }
    if as_json:
        print(json.dumps(result, indent=2, sort_keys=True))
        return
    print(f"ripwire: {version}")
    print(f"files={result['files']} inc_rows={total} in_tree={in_tree} "
          f"imports_unresolved={unresolved} -> {result['unresolved_pct']}% unresolved")


if __name__ == "__main__":
    main()
