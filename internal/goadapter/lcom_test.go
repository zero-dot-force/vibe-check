package goadapter

import (
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
)

func TestLCOM4_Cohesive(t *testing.T) {
	t.Parallel()

	pkg := loadTestPackage(t, "lcom", "cohesive")
	got := computeLCOM4(pkg)

	if got != metrics.LCOM(1) {
		t.Errorf("LCOM4: got %d, want %d", got, 1)
	}
}

func TestLCOM4_NonCohesive(t *testing.T) {
	t.Parallel()

	pkg := loadTestPackage(t, "lcom", "noncohesive")
	got := computeLCOM4(pkg)

	if got != metrics.LCOM(2) {
		t.Errorf("LCOM4: got %d, want %d", got, 2)
	}
}

func TestLCOM4_NoMethods(t *testing.T) {
	t.Parallel()

	pkg := loadTestPackage(t, "lcom", "nomethods")
	got := computeLCOM4(pkg)

	if got != metrics.LCOM(0) {
		t.Errorf("LCOM4: got %d, want %d", got, 0)
	}
}
