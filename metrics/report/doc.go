// Package report renders the deterministic architectural health report
// over [metrics.ModuleGraph] snapshots: A–F health grades, aggregate
// computation, a prioritized recommendation rule engine, byte-stable
// markdown rendering (summary and detailed modes, plus the base-vs-PR
// structural-entropy diff report), and the computed report model that
// `vibe-check report --json` serializes for machine consumers.
//
// The engine is a set of pure, table-tested functions — no I/O, no globals,
// no timestamps — so identical input renders byte-identical output, per the
// project's determinism mandate. It lives in a Layer-1 subpackage of
// metrics by design: it imports the universal model types
// ([metrics.ModuleGraph], [metrics.ModuleResult], [metrics.GraphDelta],
// [metrics.Verdict]) but depends on no adapter and no CLI code, keeping the
// engine importable by every consumer that can read a snapshot.
//
// Grades and recommendations are informational only — they advise, they
// never gate. The entropy-regression thresholds in
// [metrics.VerdictThresholds] (applied by [metrics.DecideVerdict]) remain
// the only enforcement path, and nothing in this package may be wired into
// CI gating.
package report
