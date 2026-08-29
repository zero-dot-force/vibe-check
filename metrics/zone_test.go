package metrics

import "testing"

func TestComputeZone_Classification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    Abstractness
		i    Instability
		d    Distance
		want Zone
	}{
		{
			name: "main sequence: A=0.5 I=0.5 D=0.0",
			a:    0.5,
			i:    0.5,
			d:    0.0,
			want: ZoneMainSequence,
		},
		{
			name: "main sequence boundary just inside: D=0.199",
			a:    0.5,
			i:    0.5,
			d:    0.199,
			want: ZoneMainSequence,
		},
		{
			name: "boundary D=0.2 exactly falls to zone of pain: A=0.1 I=0.1",
			a:    0.1,
			i:    0.1,
			d:    0.2,
			want: ZoneOfPain,
		},
		{
			name: "zone of pain: A=0.0 I=0.0 D=1.0",
			a:    0.0,
			i:    0.0,
			d:    1.0,
			want: ZoneOfPain,
		},
		{
			name: "zone of uselessness: A=1.0 I=1.0 D=1.0",
			a:    1.0,
			i:    1.0,
			d:    1.0,
			want: ZoneOfUselessness,
		},
		{
			name: "normal: A=0.5 I=0.3 D=0.2",
			a:    0.5,
			i:    0.3,
			d:    0.2,
			want: ZoneNormal,
		},
		{
			name: "precedence: D<0.2 takes priority over zone-of-pain",
			a:    0.1,
			i:    0.1,
			d:    0.1,
			want: ZoneMainSequence,
		},
		{
			name: "boundary A=0.2 I=0.2 is not zone-of-pain (strict <)",
			a:    0.2,
			i:    0.2,
			d:    0.5,
			want: ZoneNormal,
		},
		{
			name: "boundary A=0.8 I=0.8 is not zone-of-uselessness (strict >)",
			a:    0.8,
			i:    0.8,
			d:    0.5,
			want: ZoneNormal,
		},
		{
			name: "just inside zone-of-pain: A=0.19 I=0.19",
			a:    0.19,
			i:    0.19,
			d:    0.5,
			want: ZoneOfPain,
		},
		{
			name: "just inside zone-of-uselessness: A=0.81 I=0.81",
			a:    0.81,
			i:    0.81,
			d:    0.5,
			want: ZoneOfUselessness,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ComputeZone(tt.a, tt.i, tt.d)
			if got != tt.want {
				t.Errorf("ComputeZone(%v, %v, %v) = %v, want %v", tt.a, tt.i, tt.d, got, tt.want)
			}
		})
	}
}

func TestComputeZone_Determinism(t *testing.T) {
	t.Parallel()

	// Same inputs must produce identical outputs across multiple calls.
	// This verifies the constitutional requirement (Principle VI: Metric Fidelity)
	// that metric computations are deterministic.
	const iterations = 100
	first := ComputeZone(0.3, 0.4, 0.5)
	for i := 0; i < iterations; i++ {
		got := ComputeZone(0.3, 0.4, 0.5)
		if got != first {
			t.Fatalf("ComputeZone(0.3, 0.4, 0.5) produced non-deterministic result on iteration %d: got %v, want %v", i, got, first)
		}
	}
}
