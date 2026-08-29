package metrics

import "math"

// ComputeInstability computes I = Ce / (Ca + Ce).
// When both Ca and Ce are 0, returns 0.0 (maximally stable by convention).
// This follows the convention that an isolated module with no coupling
// relationships is treated as stable rather than unstable.
func ComputeInstability(ca, ce int) Instability {
	denom := ca + ce
	if denom == 0 {
		return 0.0
	}
	return Instability(float64(ce) / float64(denom))
}

// ComputeAbstractness computes A = abstractTypes / totalExported.
// When totalExported is 0, returns 0.0 (a module with no exported types
// is treated as fully concrete).
func ComputeAbstractness(abstractTypes, totalExported int) Abstractness {
	if totalExported == 0 {
		return 0.0
	}
	return Abstractness(float64(abstractTypes) / float64(totalExported))
}

// ComputeDistance computes D = |A + I - 1|.
// The result measures how far a module is from the main sequence, where
// the main sequence represents the ideal balance between abstractness
// and instability.
func ComputeDistance(a Abstractness, i Instability) Distance {
	return Distance(math.Abs(float64(a) + float64(i) - 1.0))
}

// ComputeZone classifies a module's position relative to the main sequence.
// Precedence (evaluated in order):
//  1. main-sequence: D < 0.2
//  2. zone-of-pain: A < 0.2 AND I < 0.2
//  3. zone-of-uselessness: A > 0.8 AND I > 0.8
//  4. normal: all other cases
func ComputeZone(a Abstractness, i Instability, d Distance) Zone {
	if d < 0.2 {
		return ZoneMainSequence
	}
	if a < 0.2 && i < 0.2 {
		return ZoneOfPain
	}
	if a > 0.8 && i > 0.8 {
		return ZoneOfUselessness
	}
	return ZoneNormal
}
