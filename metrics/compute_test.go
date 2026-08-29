package metrics

import (
	"math"
	"testing"
)

const epsilon = 1e-10

// assertFloatEq compares two float64 values within an epsilon tolerance
// and reports a test failure with a descriptive message if they differ.
func assertFloatEq(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) >= epsilon {
		t.Errorf("got %v, want %v (within epsilon %v)", got, want, epsilon)
	}
}

func TestComputeInstability_HappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ca   int
		ce   int
		want float64
	}{
		{name: "mixed coupling Ca=3 Ce=7", ca: 3, ce: 7, want: 0.7},
		{name: "equal coupling Ca=5 Ce=5", ca: 5, ce: 5, want: 0.5},
		{name: "Ca=1 Ce=3", ca: 1, ce: 3, want: 0.75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ComputeInstability(tt.ca, tt.ce)
			assertFloatEq(t, float64(got), tt.want)
		})
	}
}

func TestComputeInstability_ZeroDenominator(t *testing.T) {
	t.Parallel()

	got := ComputeInstability(0, 0)
	assertFloatEq(t, float64(got), 0.0)
}

func TestComputeInstability_MaximallyStable(t *testing.T) {
	t.Parallel()

	got := ComputeInstability(5, 0)
	assertFloatEq(t, float64(got), 0.0)
}

func TestComputeInstability_MaximallyUnstable(t *testing.T) {
	t.Parallel()

	got := ComputeInstability(0, 5)
	assertFloatEq(t, float64(got), 1.0)
}

func TestComputeInstability_Determinism(t *testing.T) {
	t.Parallel()

	// Same inputs must produce identical outputs across multiple calls.
	// This verifies the constitutional requirement (Principle VI: Metric Fidelity)
	// that metric computations are deterministic.
	const iterations = 100
	first := ComputeInstability(3, 7)
	for i := 0; i < iterations; i++ {
		got := ComputeInstability(3, 7)
		if got != first {
			t.Fatalf("ComputeInstability(3, 7) produced non-deterministic result on iteration %d: got %v, want %v", i, got, first)
		}
	}
}

func TestComputeAbstractness_FullyAbstract(t *testing.T) {
	t.Parallel()

	got := ComputeAbstractness(3, 3)
	assertFloatEq(t, float64(got), 1.0)
}

func TestComputeAbstractness_FullyConcrete(t *testing.T) {
	t.Parallel()

	got := ComputeAbstractness(0, 5)
	assertFloatEq(t, float64(got), 0.0)
}

func TestComputeAbstractness_Mixed(t *testing.T) {
	t.Parallel()

	got := ComputeAbstractness(2, 5)
	assertFloatEq(t, float64(got), 0.4)
}

func TestComputeAbstractness_NoExports(t *testing.T) {
	t.Parallel()

	got := ComputeAbstractness(0, 0)
	assertFloatEq(t, float64(got), 0.0)
}

func TestComputeAbstractness_Determinism(t *testing.T) {
	t.Parallel()

	const iterations = 100
	first := ComputeAbstractness(2, 5)
	for i := 0; i < iterations; i++ {
		got := ComputeAbstractness(2, 5)
		if got != first {
			t.Fatalf("ComputeAbstractness(2, 5) produced non-deterministic result on iteration %d: got %v, want %v", i, got, first)
		}
	}
}

func TestComputeDistance_OnMainSequence(t *testing.T) {
	t.Parallel()

	got := ComputeDistance(0.5, 0.5)
	assertFloatEq(t, float64(got), 0.0)
}

func TestComputeDistance_ZoneOfPain(t *testing.T) {
	t.Parallel()

	// Zone of pain: concrete (A=0.0) and stable (I=0.0).
	// D = |0.0 + 0.0 - 1| = 1.0
	got := ComputeDistance(0.0, 0.0)
	assertFloatEq(t, float64(got), 1.0)
}

func TestComputeDistance_ZoneOfUselessness(t *testing.T) {
	t.Parallel()

	// Zone of uselessness: abstract (A=1.0) and unstable (I=1.0).
	// D = |1.0 + 1.0 - 1| = 1.0
	got := ComputeDistance(1.0, 1.0)
	assertFloatEq(t, float64(got), 1.0)
}

func TestComputeDistance_Determinism(t *testing.T) {
	t.Parallel()

	const iterations = 100
	first := ComputeDistance(0.3, 0.4)
	for i := 0; i < iterations; i++ {
		got := ComputeDistance(0.3, 0.4)
		if got != first {
			t.Fatalf("ComputeDistance(0.3, 0.4) produced non-deterministic result on iteration %d: got %v, want %v", i, got, first)
		}
	}
}

func TestComputeAbstractness_AbstractExceedsTotalExported(t *testing.T) {
	t.Parallel()

	// When abstractTypes exceeds totalExported, the function trusts adapter
	// input and may produce values > 1.0. This documents the behavior for
	// invalid inputs — the adapter is responsible for providing valid data.
	got := ComputeAbstractness(5, 3)
	if float64(got) <= 1.0 {
		t.Errorf("ComputeAbstractness(5, 3) = %v, expected > 1.0 for invalid input", got)
	}
}

func TestComputeInstability_NegativeInputs(t *testing.T) {
	t.Parallel()

	// Negative coupling counts are semantically invalid. The function does
	// not guard against them — callers are responsible for providing
	// non-negative values. This test documents the unguarded behavior.
	got := ComputeInstability(-1, 3)
	if float64(got) < 0.0 || float64(got) > 1.0 {
		// Document that negative inputs produce out-of-range values.
		t.Logf("ComputeInstability(-1, 3) = %v (out of [0.0, 1.0] range for invalid input)", got)
	}
}
