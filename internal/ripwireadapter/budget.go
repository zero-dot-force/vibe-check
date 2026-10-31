package ripwireadapter

import (
	"fmt"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// warningErrorBudget is the warning code for the spike-measured error
// budgets of the ripwire-backed coupling data, per language.
const warningErrorBudget = "error-budget"

// errorBudgetWarning returns the language's spike-measured error budget
// warning (openspec/changes/ripwire-adapter spike #67): the measured
// inaccuracy of sourcing coupling data from ripwire's name-based,
// file-level include scan. Python's budget combines lost edges (imports
// the scan misses) with spurious edges (includes it over-credits);
// TypeScript's budget is the unresolved-target rate.
func errorBudgetWarning(language string) metrics.Warning {
	switch language {
	case languagePython:
		return metrics.Warning{
			Code: warningErrorBudget,
			Message: "coupling data sourced from ripwire's name-based scan carries a " +
				"spike-measured error budget: 1.9% of edges lost, 2.2% spurious",
		}
	case languageTypeScript:
		return metrics.Warning{
			Code: warningErrorBudget,
			Message: "coupling data sourced from ripwire's path-based scan carries a " +
				"spike-measured error budget: 0.16% of targets unresolved",
		}
	default:
		return metrics.Warning{
			Code:    warningErrorBudget,
			Message: fmt.Sprintf("coupling data sourced from ripwire for language %q has no measured error budget", language),
		}
	}
}

// graphStatusForWarnings implements the assembly rule that any warning on
// a graph forces status partial: a warning means a caveat affects metric
// accuracy, so the graph must never present as fully trustworthy
// (metrics.StatusComplete). A warning-free graph is complete.
func graphStatusForWarnings(warnings []metrics.Warning) metrics.Status {
	if len(warnings) > 0 {
		return metrics.StatusPartial
	}
	return metrics.StatusComplete
}
