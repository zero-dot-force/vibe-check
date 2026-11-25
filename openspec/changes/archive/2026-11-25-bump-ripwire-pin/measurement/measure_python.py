#!/usr/bin/env python3
"""Measure ripwire's Python coupling accuracy against an ast ground truth.

openspec/changes/bump-ripwire-pin design D6 / tasks 5.1-5.2. Spec-phase
evidence, never built or run by CI. Run with `python3 -I` (isolated mode)
so nothing in the measured tree can shadow a standard-library module.

    python3 -I measure_python.py <ripwire-binary> <crawl-root> [--json] [--tracked-build]

<crawl-root> is the directory ripwire crawls (for a src-layout project, its
src/ directory), exactly as the adapter would choose it.

Ground truth: every .py file under the crawl root is parsed with ast (never
imported or executed), and every import statement is resolved to in-tree
files with Python's own rules:

  * `import a.b.c [as x]` -> the file of the deepest in-tree module among
    a.b.c, a.b, a (a/b/c.py or a/b/c/__init__.py).
  * `from X import n` (X absolute or relative) -> X's file (the package or
    module imported from always loads) plus X.n's file when X.n is an
    in-tree module or package (a member edge): the "wide form" upstream
    adopted in redhat-et/ripwire#533. `*` -> X only.
  * imports anywhere in the file count (module level, functions, classes,
    TYPE_CHECKING blocks): the static import graph, as ripwire reads it.
  * absolute names resolve first from the crawl root, then from the
    importer's package-chain root (the nearest ancestor directory holding
    no __init__.py) — the sys.path entry a script or test runner adds when
    it runs a file in that tree (Django's tests/, scripts/). This matches
    the fallback ripwire adopted with its src-layout fix (Train 44).
  * targets outside the tree (stdlib, third-party) are not edges.
  * directories ripwire's crawl denylist prunes are excluded from both
    sides: their absence is disclosed separately (denylist-pruned warning),
    not an accuracy error. With --tracked-build (ripwire >= 0.7.0), a
    build/ directory under which git tracks files is crawled, and only its
    tracked files count: the rule ripwire 0.7.0 adopted ("a build/ directory
    whose files git tracks is indexed"). Without it, build/ is always
    pruned, as in 0.6.9 and 0.6.10.

Measured (ripwire side):

  * directory grain, what vibe-check consumes: `--arch` <m> rows ca/ce vs
    ground-truth Ca/Ce per directory module -> % exact and mean absolute
    error. This is the Phase 0 spike's headline metric (pip @ 1000636 on
    0.6.9: Ca 81% exact / MAE 0.19, Ce 85% / MAE 0.21).
  * file grain, count-based: `--export=cc.json` per-file fan_out vs the
    ground-truth per-file out-degree. lost = sum(max(0, gt - rw)) / sum(gt);
    spurious = sum(max(0, rw - gt)) / sum(gt). A file that both loses and
    gains an edge nets to zero, so both figures are LOWER BOUNDS on the true
    edge-level rates. (The spike reconstructed edge identities from --deps
    <inc t=> text; since redhat-et/ripwire#533 one written import can
    resolve to two files, so that reconstruction no longer holds.)
"""

import ast
import json
import os
import subprocess
import sys
import xml.etree.ElementTree as ET

# ripwire's crawl denylist (pinned release): directories pruned by name.
DENYLIST = {"build", "dist", "out", "target", "vendor", "third_party", "captures"}


def git_tracked(root):
    """Crawl-root-relative paths git tracks under root (empty outside git)."""
    out = subprocess.run(["git", "-C", root, "ls-files", "-z"], capture_output=True)
    if out.returncode != 0:
        return set()
    return {p for p in out.stdout.decode("utf-8", "surrogateescape").split("\0") if p}


def python_files(root, pruned=None, tracked=None):
    """Yield crawl-root-relative paths of .py files, pruning like ripwire.

    When pruned is a set, the crawl-root-relative paths of the denylisted
    directories that were skipped are added to it. When tracked is a set of
    git-tracked paths (--tracked-build), a build/ directory holding any of
    them is crawled, keeping only tracked files beneath it.
    """
    tracked_build = set()  # crawl-root-relative build/ dirs being crawled
    for dirpath, dirnames, filenames in os.walk(root):
        rel_dir = os.path.relpath(dirpath, root).replace(os.sep, "/")
        rel_dir = "" if rel_dir == "." else rel_dir + "/"
        inside_build = any(rel_dir.startswith(b + "/") for b in tracked_build)
        kept = []
        for d in sorted(dirnames):
            rel = rel_dir + d
            if d == "build" and tracked is not None and any(t.startswith(rel + "/") for t in tracked):
                tracked_build.add(rel)
                kept.append(d)
                continue
            if d in DENYLIST and pruned is not None:
                pruned.add(rel)
            if d not in DENYLIST and not d.startswith("."):
                kept.append(d)
        dirnames[:] = kept
        for name in sorted(filenames):
            rel = rel_dir + name
            if name.endswith(".py") and (not inside_build or rel in tracked):
                yield rel


def module_name(rel):
    """Dotted module name for a crawl-root-relative .py path."""
    parts = rel[:-3].split("/")
    if parts[-1] == "__init__":
        parts = parts[:-1]
    return ".".join(parts)


def ground_truth(root, tracked=None):
    """Return (files, edges): in-tree files and resolved (src, dst) file edges."""
    pruned = set()
    files = list(python_files(root, pruned, tracked))
    by_module = {module_name(f): f for f in files}
    file_set = set(files)

    def chain_root(rel):
        """Nearest ancestor directory (crawl-root-relative) without __init__.py."""
        parts = rel.split("/")[:-1]
        while parts and "/".join(parts + ["__init__.py"]) in file_set:
            parts.pop()
        return "/".join(parts)

    def lookup(dotted, prefix):
        """File for a dotted module name resolved under a root prefix."""
        if not prefix:
            return by_module.get(dotted)
        base = prefix + "/" + dotted.replace(".", "/")
        for candidate in (base + ".py", base + "/__init__.py"):
            if candidate in file_set:
                return candidate
        return None
    edges = set()
    parse_errors = []

    for rel in files:
        try:
            with open(os.path.join(root, rel), "rb") as handle:
                tree = ast.parse(handle.read(), filename=rel)
        except (SyntaxError, ValueError) as exc:
            parse_errors.append(f"{rel}: {exc}")
            continue
        is_package = rel.endswith("/__init__.py") or rel == "__init__.py"
        own = module_name(rel)
        package = own if is_package else own.rpartition(".")[0]

        roots = [""]
        extra = chain_root(rel)
        if extra:
            roots.append(extra)

        def find(dotted):
            for prefix in roots:
                hit = lookup(dotted, prefix)
                if hit:
                    return hit
            return None

        def into_pruned(dotted):
            """True when dotted names a path inside a denylist-pruned directory."""
            for prefix in roots:
                path = (prefix + "/" if prefix else "") + dotted.replace(".", "/")
                if any(path == d or path.startswith(d + "/") for d in pruned):
                    return True
            return False

        def deepest(dotted):
            if into_pruned(dotted):
                # The target was pruned from the crawl: excluded from both
                # sides (disclosed by denylist-pruned), never attributed to
                # its parent package.
                return None
            parts = dotted.split(".")
            for i in range(len(parts), 0, -1):
                hit = find(".".join(parts[:i]))
                if hit:
                    return hit
            return None

        for node in ast.walk(tree):
            targets = []
            if isinstance(node, ast.Import):
                for alias in node.names:
                    targets.append(deepest(alias.name))
            elif isinstance(node, ast.ImportFrom):
                if node.level:
                    anchor = package.split(".") if package else []
                    up = node.level - 1
                    if up > len(anchor):
                        continue  # relative import escapes the tree
                    base_parts = anchor[: len(anchor) - up] if up else anchor
                    if node.module:
                        base_parts = base_parts + node.module.split(".")
                    base = ".".join(base_parts)
                else:
                    base = node.module or ""
                relative = node.level > 0
                resolve = by_module.get if relative else find
                if base:
                    # The package (or module) imported from always loads.
                    targets.append(resolve(base) or (None if relative else deepest(base)))
                for alias in node.names:
                    member = f"{base}.{alias.name}" if base else alias.name
                    if alias.name != "*":
                        targets.append(resolve(member))
            for dst in targets:
                if dst and dst != rel:
                    edges.add((rel, dst))
    return files, edges, parse_errors


def run(ripwire, root, *args):
    out = subprocess.run(
        [ripwire, ".", *args, "--no-cache"],
        cwd=root, capture_output=True, check=True,
    )
    return out.stdout


def ripwire_dirs(ripwire, root, rules):
    xml = run(ripwire, root, f"--arch={rules}")
    arch = ET.fromstring(xml.decode())
    rows = {}
    for m in arch.iter("m"):
        path = m.get("path").removeprefix("./")
        rows[path] = (int(m.get("ca", 0)), int(m.get("ce", 0)), m.get("zone"))
    return rows


def ripwire_fan_out(ripwire, root):
    data = json.loads(run(ripwire, root, "--export=cc.json"))
    fan = {}

    def walk(node, prefix):
        path = f"{prefix}/{node['name']}" if prefix else node["name"]
        if node.get("type") == "File":
            fan[path.split("/", 1)[1]] = node["attributes"].get("fan_out", 0)
        for child in node.get("children", []):
            walk(child, path)

    for node in data["nodes"]:
        walk(node, "")
    return fan


def directory_metrics(files, edges):
    dirs = {os.path.dirname(f) or "." for f in files}
    ca = {d: set() for d in dirs}
    ce = {d: set() for d in dirs}
    for src, dst in edges:
        a, b = os.path.dirname(src) or ".", os.path.dirname(dst) or "."
        if a != b:
            ce[a].add(b)
            ca[b].add(a)
    return {d: (len(ca[d]), len(ce[d])) for d in dirs}


def main():
    if len(sys.argv) < 3:
        sys.exit(__doc__)
    ripwire, root = sys.argv[1], os.path.abspath(sys.argv[2])
    as_json = "--json" in sys.argv[3:]
    tracked = git_tracked(root) if "--tracked-build" in sys.argv[3:] else None
    here = os.path.dirname(os.path.abspath(__file__))
    rules = os.path.join(here, "arch-rules.txt")

    version = subprocess.run([ripwire, "--version"], capture_output=True, text=True, check=True).stdout.strip()
    files, edges, parse_errors = ground_truth(root, tracked)
    gt_dirs = directory_metrics(files, edges)
    rw_dirs = ripwire_dirs(ripwire, root, rules)
    rw_fan = ripwire_fan_out(ripwire, root)

    # Directory grain: compare on directories holding ground-truth files.
    compared = sorted(d for d in gt_dirs if d in rw_dirs)
    missing = sorted(d for d in gt_dirs if d not in rw_dirs)
    ca_exact = sum(1 for d in compared if rw_dirs[d][0] == gt_dirs[d][0])
    ce_exact = sum(1 for d in compared if rw_dirs[d][1] == gt_dirs[d][1])
    ca_mae = sum(abs(rw_dirs[d][0] - gt_dirs[d][0]) for d in compared) / max(1, len(compared))
    ce_mae = sum(abs(rw_dirs[d][1] - gt_dirs[d][1]) for d in compared) / max(1, len(compared))
    unknown = sum(1 for d in compared if rw_dirs[d][2] == "unknown")

    # File grain, count-based.
    gt_out = {f: 0 for f in files}
    for src, _ in edges:
        gt_out[src] += 1
    total = sum(gt_out.values())
    lost = sum(max(0, gt_out[f] - rw_fan.get(f, 0)) for f in files)
    spurious = sum(max(0, rw_fan.get(f, 0) - gt_out[f]) for f in files)

    worst = sorted(
        compared,
        key=lambda d: -(abs(rw_dirs[d][0] - gt_dirs[d][0]) + abs(rw_dirs[d][1] - gt_dirs[d][1])),
    )[:10]
    result = {
        "ripwire": version,
        "crawl_root": root,
        "files": len(files),
        "parse_errors": parse_errors,
        "gt_edges": total,
        "dirs_compared": len(compared),
        "dirs_missing_from_arch": missing,
        "dirs_zone_unknown": unknown,
        "ca_exact_pct": round(100 * ca_exact / max(1, len(compared)), 1),
        "ca_mae": round(ca_mae, 2),
        "ce_exact_pct": round(100 * ce_exact / max(1, len(compared)), 1),
        "ce_mae": round(ce_mae, 2),
        "file_lost_pct_lower_bound": round(100 * lost / max(1, total), 2),
        "file_spurious_pct_lower_bound": round(100 * spurious / max(1, total), 2),
        "worst_dirs": [
            {"dir": d, "gt_ca_ce": gt_dirs[d], "rw_ca_ce": rw_dirs[d][:2]} for d in worst
            if gt_dirs[d] != rw_dirs[d][:2]
        ],
    }
    if as_json:
        print(json.dumps(result, indent=2, sort_keys=True))
        return
    print(f"ripwire: {version}")
    print(f"files={result['files']} gt_edges={total} dirs={len(compared)} "
          f"missing_from_arch={len(missing)} zone_unknown={unknown} parse_errors={len(parse_errors)}")
    print(f"Ca exact {result['ca_exact_pct']}% (MAE {result['ca_mae']})  "
          f"Ce exact {result['ce_exact_pct']}% (MAE {result['ce_mae']})")
    print(f"file edges (count-based lower bounds): lost {result['file_lost_pct_lower_bound']}%  "
          f"spurious {result['file_spurious_pct_lower_bound']}%")
    for w in result["worst_dirs"]:
        print(f"  {w['dir']}: gt ca/ce {w['gt_ca_ce']} vs ripwire {w['rw_ca_ce']}")


if __name__ == "__main__":
    main()
