package goadapter

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// Extension capability identifiers for Go-specific metrics. These are optional
// capabilities beyond the universal metrics.Cap* set defined in the metrics
// package. They live in the adapter package because they are language-specific,
// and are surfaced via [Adapter.ExtensionCapabilities] rather than the core
// [Adapter.Capabilities]. Each constant also serves as the key under which its
// value is stored in a ModuleResult's Extensions map.
const (
	// CapInterfaceWidth is the extension capability and Extensions map key for
	// per-interface flattened method counts (map[string]int).
	CapInterfaceWidth = "go.interfaceWidth"
	// CapInterfaceProximity is the extension capability and Extensions map key
	// for per-interface "producer"/"consumer" classification (map[string]string).
	CapInterfaceProximity = "go.interfaceProximity"
)

// computeExtensions computes Go-specific extension metrics for a package.
// It populates "go.interfaceWidth" (method count per exported interface) and
// "go.interfaceProximity" ("producer" or "consumer" per interface).
//
// Returns nil if no exported interfaces exist in the package, which causes
// the extensions field to be omitted from JSON output (omitempty).
func computeExtensions(pkg *packages.Package) map[string]any {
	if pkg.Types == nil {
		return nil
	}

	scope := pkg.Types.Scope()
	widths := make(map[string]int)
	proximities := make(map[string]string)

	// Collect all exported interfaces and their flattened method counts.
	var ifaceTypes []*types.Interface
	var ifaceNames []string

	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		tn, ok := obj.(*types.TypeName)
		if !ok || !tn.Exported() {
			continue
		}

		iface, ok := tn.Type().Underlying().(*types.Interface)
		if !ok {
			continue
		}

		// Flattened method count includes methods from embedded interfaces.
		widths[name] = iface.NumMethods()
		ifaceTypes = append(ifaceTypes, iface)
		ifaceNames = append(ifaceNames, name)
	}

	if len(widths) == 0 {
		return nil
	}

	// Collect all concrete types in the package for proximity analysis.
	var concreteTypes []types.Type
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		tn, ok := obj.(*types.TypeName)
		if !ok || !tn.Exported() {
			continue
		}

		// Skip interfaces — we want concrete types only.
		if _, isIface := tn.Type().Underlying().(*types.Interface); isIface {
			continue
		}

		// Check both the type and its pointer variant.
		concreteTypes = append(concreteTypes, tn.Type())
	}

	// Determine proximity for each interface.
	for i, iface := range ifaceTypes {
		name := ifaceNames[i]
		proximities[name] = computeProximity(iface, concreteTypes)
	}

	return map[string]any{
		CapInterfaceWidth:     widths,
		CapInterfaceProximity: proximities,
	}
}

// computeProximity determines whether an interface is a "producer" or "consumer"
// in the context of the package where it is declared.
//
// An interface is a "producer" if any concrete type in the same package implements
// it (the package produces implementations). Otherwise it is a "consumer" (the
// package consumes implementations provided by other packages).
func computeProximity(iface *types.Interface, concreteTypes []types.Type) string {
	for _, ct := range concreteTypes {
		// Check if the concrete type or its pointer implements the interface.
		if types.Implements(ct, iface) {
			return "producer"
		}
		ptr := types.NewPointer(ct)
		if types.Implements(ptr, iface) {
			return "producer"
		}
	}
	return "consumer"
}

// InterfaceWidths extracts the "go.interfaceWidth" extension from a
// [metrics.ModuleResult]'s Extensions map. It handles the JSON round-trip
// conversion where int values become float64 after unmarshaling.
//
// Returns an error if the key is missing or has an unexpected type.
func InterfaceWidths(extensions map[string]any) (map[string]int, error) {
	raw, ok := extensions[CapInterfaceWidth]
	if !ok {
		return nil, fmt.Errorf("extension key %q not found", CapInterfaceWidth)
	}

	// Before JSON round-trip: map[string]int.
	if typed, ok := raw.(map[string]int); ok {
		result := make(map[string]int, len(typed))
		for k, v := range typed {
			result[k] = v
		}
		return result, nil
	}

	// After JSON round-trip: map[string]interface{} with float64 values.
	rawMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("extension key %q: expected map, got %T", CapInterfaceWidth, raw)
	}

	result := make(map[string]int, len(rawMap))
	for k, v := range rawMap {
		switch n := v.(type) {
		case float64:
			result[k] = int(n)
		case int:
			result[k] = n
		default:
			return nil, fmt.Errorf("extension key %q: value for %q has unexpected type %T", CapInterfaceWidth, k, v)
		}
	}

	return result, nil
}

// InterfaceProximities extracts the "go.interfaceProximity" extension from a
// [metrics.ModuleResult]'s Extensions map.
//
// Returns an error if the key is missing or has an unexpected type.
func InterfaceProximities(extensions map[string]any) (map[string]string, error) {
	raw, ok := extensions[CapInterfaceProximity]
	if !ok {
		return nil, fmt.Errorf("extension key %q not found", CapInterfaceProximity)
	}

	// Before JSON round-trip: map[string]string.
	if typed, ok := raw.(map[string]string); ok {
		result := make(map[string]string, len(typed))
		for k, v := range typed {
			result[k] = v
		}
		return result, nil
	}

	// After JSON round-trip: map[string]interface{} with string values.
	rawMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("extension key %q: expected map, got %T", CapInterfaceProximity, raw)
	}

	result := make(map[string]string, len(rawMap))
	for k, v := range rawMap {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("extension key %q: value for %q has unexpected type %T", CapInterfaceProximity, k, v)
		}
		result[k] = s
	}

	return result, nil
}
