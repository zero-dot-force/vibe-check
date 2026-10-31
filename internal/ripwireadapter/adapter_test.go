// Adapter scaffold tests — openspec/changes/ripwire-adapter task 2.1.
//
// They pin the ripwire-adapter spec's adapter-instance and per-language
// capability gating scenarios for the production scaffold: the exact language
// identifier each constructor reports, the exact seven-capability set per
// language (duplication deliberately undeclared — it becomes an
// "unsupported-metric" warning at analysis time), and the distinctness of the
// two constructors' adapters. Analyze is a not-implemented stub at this stage
// (tasks 2.2, 2.3, 4.x, and 6.2 build it out), so it carries no tests yet.

package ripwireadapter

import (
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// adapterConstructors enumerates every public constructor with the exact
// language identifier its adapter must report, so the identifier and
// capability contract tests run against both language instances.
func adapterConstructors() []struct {
	name string
	new  func() *Adapter
	want string
} {
	return []struct {
		name string
		new  func() *Adapter
		want string
	}{
		{name: "python", new: NewPython, want: "python"},
		{name: "typescript", new: NewTypeScript, want: "typescript"},
	}
}

// wantCapabilities returns the exact capability set every ripwire adapter
// must declare: the seven universal metrics the hybrid pipeline sources. It
// is expressed directly in metrics constants so the expected set stays pinned
// in the test's own terms, never derived from production code.
func wantCapabilities() []metrics.Capability {
	return []metrics.Capability{
		metrics.CapAfferentCoupling,
		metrics.CapEfferentCoupling,
		metrics.CapInstability,
		metrics.CapAbstractness,
		metrics.CapDistance,
		metrics.CapLCOM,
		metrics.CapCircularDeps,
	}
}

// TestLanguage_IdentifierPerConstructor asserts the exact language identifier
// each constructor's adapter reports (spec scenario "Language identifiers").
// The identifiers double as metrics.Registry keys, so drift here breaks
// adapter selection.
func TestLanguage_IdentifierPerConstructor(t *testing.T) {
	t.Parallel()

	for _, tc := range adapterConstructors() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := tc.new()
			if got := a.Language(); got != tc.want {
				t.Errorf("Language() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestCapabilities_ExactSetPerLanguage asserts the exact capability set per
// language (spec scenario "Capability set per language"): exactly the seven
// universal capabilities, with no duplicate declarations and no duplication
// capability — duplication is unsupported for both languages and must surface
// as an "unsupported-metric" warning at analysis time, never as a declared
// capability.
func TestCapabilities_ExactSetPerLanguage(t *testing.T) {
	t.Parallel()

	want := wantCapabilities()

	for _, tc := range adapterConstructors() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.new().Capabilities()

			if len(got) != len(want) {
				t.Errorf("Capabilities() declares %d capabilities, want %d: %v", len(got), len(want), got)
			}

			// Count occurrences rather than checking membership alone:
			// membership plus total length cannot catch a duplicate that
			// displaces a missing entry.
			counts := make(map[metrics.Capability]int, len(want))
			for _, c := range got {
				counts[c]++
			}
			if len(counts) != len(want) {
				t.Errorf("Capabilities() declares only %d distinct capabilities, want %d — a capability is duplicated: %v", len(counts), len(want), got)
			}

			for _, w := range want {
				if n := counts[w]; n != 1 {
					t.Errorf("Capabilities() declares %q %d times, want exactly once", w, n)
				}
			}

			// Exact-set equality above already forbids it, but the
			// deliberate omission of duplication is a spec requirement of
			// its own, so assert it by name.
			if n := counts[metrics.CapDuplication]; n != 0 {
				t.Errorf("Capabilities() must not declare %q (%d times): duplication is unsupported and becomes an unsupported-metric warning at analysis time", metrics.CapDuplication, n)
			}
		})
	}
}

// TestNew_ConstructorsReturnDistinctAdapters asserts that NewPython and
// NewTypeScript produce distinct adapters — different language identifiers
// and different instances — so each registers independently under its own
// metrics.Registry key (spec scenario "Registry registration and
// retrieval"). The registry keys adapters by Language() and rejects a second
// adapter for the same key, so any instance sharing between the two
// constructors would break adapter selection.
func TestNew_ConstructorsReturnDistinctAdapters(t *testing.T) {
	t.Parallel()

	py := NewPython()
	ts := NewTypeScript()

	if py.Language() == ts.Language() {
		t.Errorf("NewPython() and NewTypeScript() both report language %q; each constructor must yield its own language identifier", py.Language())
	}
	if py == ts {
		t.Errorf("NewPython() and NewTypeScript() returned the same adapter instance; each language must get its own adapter")
	}

	// The consumer contract for distinctness: both adapters register under
	// their own keys and each retrieval returns an adapter reporting its own
	// identifier.
	reg := metrics.NewRegistry()
	if err := reg.Register(py); err != nil {
		t.Fatalf("register python adapter: %v", err)
	}
	if err := reg.Register(ts); err != nil {
		t.Fatalf("register typescript adapter: %v", err)
	}

	gotPy, err := reg.Get("python")
	if err != nil {
		t.Fatalf("get python adapter: %v", err)
	}
	if got := gotPy.Language(); got != "python" {
		t.Errorf("registry Get(%q) returned an adapter reporting Language() %q, want %q", "python", got, "python")
	}

	gotTS, err := reg.Get("typescript")
	if err != nil {
		t.Fatalf("get typescript adapter: %v", err)
	}
	if got := gotTS.Language(); got != "typescript" {
		t.Errorf("registry Get(%q) returned an adapter reporting Language() %q, want %q", "typescript", got, "typescript")
	}
}
