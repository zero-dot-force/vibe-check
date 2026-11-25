package ripwireadapter

import (
	"fmt"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// warningErrorBudget is the warning code for the measured error budgets of
// the ripwire-backed coupling data, per language.
const warningErrorBudget = "error-budget"

// errorBudgetWarning returns the language's measured error-budget warning:
// the inaccuracy of sourcing coupling data from ripwire's name-based,
// file-level import scan, measured against the pinned ripwire release by
// the harness in openspec/changes/archive/2026-11-25-bump-ripwire-pin/measurement (results.md
// there). Python's budget states directory-grain Ca/Ce exactness on pip and
// django plus count-based file-edge loss and gain (lower bounds);
// TypeScript's budget is ripwire's disclosed unresolved-import rate on
// rollup. The message names the pinned version so a reader can tell which
// instrument the figures describe.
func errorBudgetWarning(language string) metrics.Warning {
	switch language {
	case languagePython:
		return metrics.Warning{
			Code: warningErrorBudget,
			Message: fmt.Sprintf("coupling data sourced from ripwire %s's name-based scan carries a "+
				"measured error budget (pip, django): directory Ca/Ce 100%% exact; file edges "+
				"at least 0.11%% lost and 0.01%% spurious (count-based lower bounds)", pinnedRipwireVersion),
		}
	case languageTypeScript:
		return metrics.Warning{
			Code: warningErrorBudget,
			Message: fmt.Sprintf("coupling data sourced from ripwire %s's path-based scan carries a "+
				"measured error budget (rollup): 0.16%% of import targets unresolved", pinnedRipwireVersion),
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
