package ripwireadapter

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// cohesionMethod is one method node of the LCOM4 graph: the method name
// and the set of instance fields its body touches.
type cohesionMethod struct {
	name   string
	fields map[string]bool
}

// cohesionClass is one class declaration with its method nodes.
type cohesionClass struct {
	name    string
	methods []cohesionMethod
}

// pyDefRegex matches a Python function declaration line (plain or async).
var pyDefRegex = regexp.MustCompile(`^(async\s+)?def\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)

// pySelfFieldRegex extracts instance-field references from a Python line.
var pySelfFieldRegex = regexp.MustCompile(`self\.([A-Za-z_][A-Za-z0-9_]*)`)

// tsClassRegex matches a TypeScript class declaration.
var tsClassRegex = regexp.MustCompile(`\bclass\s+([A-Za-z_$][A-Za-z0-9_$]*)`)

// tsMethodRegex matches a TypeScript method member declaration inside a
// class body: optional modifiers, a name, optional type parameters, a
// parenthesized parameter list, an optional return type, and an opening
// brace on the same line. Property declarations (which end in ";" or
// carry "=" before the parameters) and getters/setters (two tokens
// before the parameters) do not match.
var tsMethodRegex = regexp.MustCompile(
	`^\s*(?:(?:public|private|protected|static|readonly|async|override|abstract)\s+)*` +
		`([A-Za-z_$][A-Za-z0-9_$]*)\s*(?:<[^>]*>)?\s*\([^)]*\)\s*(?::[^{]*)?\{`)

// tsThisFieldRegex extracts instance-field references from a TypeScript line.
var tsThisFieldRegex = regexp.MustCompile(`this\.([A-Za-z_$][A-Za-z0-9_$]*)`)

// lcom4Scan walks the crawl root's sources for the given language and
// returns, per root-relative directory, the module's LCOM4 value: the
// sum over the directory's class declarations of the connected-component
// count of the class's method graph (methods sharing an instance-field
// reference are connected; a class with no bound methods contributes
// zero). Module-level functions are excluded — they are not methods and
// carry no instance fields.
//
// The node definition is deliberately broader than the Go adapter's
// exported-only filter: underscore-prefixed Python classes count, because
// ripwire's coupling data already covers non-exported code. The parses
// are light (indent stack for Python, brace depth for TypeScript) with
// documented limitations: Python base lists and docstring prose are
// scanned as plain lines; TypeScript strings and template literals
// containing braces can skew the depth count; arrow-function class
// properties and getters/setters are not method nodes.
//
// Directories ripwire prunes (the shared denylist) are skipped, matching
// the coupling data's view. Degradation follows task 5.3: unreadable
// files and directories produce parse-error warnings and contribute
// zeroed metrics rather than aborting the scan; only a missing or
// non-directory crawl root, or an unsupported language, fails closed.
func lcom4Scan(crawlRoot, language string) (map[string]int, []metrics.Warning, error) {
	extensions := map[string]bool{}
	switch language {
	case languagePython:
		extensions[".py"] = true
	case languageTypeScript:
		extensions[".ts"] = true
		extensions[".tsx"] = true
	default:
		return nil, nil, fmt.Errorf("lcom4 scan: unsupported language %q", language)
	}

	info, err := os.Stat(crawlRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("lcom4 scan: stat %s: %w", crawlRoot, err)
	}
	if !info.IsDir() {
		return nil, nil, fmt.Errorf("lcom4 scan: %s: %w", crawlRoot, errNotDirectory)
	}

	deny := make(map[string]bool, len(crawlDenylistNames))
	for _, name := range crawlDenylistNames {
		deny[name] = true
	}

	var warnings []metrics.Warning
	results := make(map[string]int)
	err = filepath.WalkDir(crawlRoot, func(filePath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			warnings = append(warnings, unreadablePathWarning(crawlRoot, filePath, walkErr))
			return nil // skip the unreadable subtree
		}
		if d.IsDir() {
			if filePath != crawlRoot && deny[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if !extensions[strings.ToLower(filepath.Ext(d.Name()))] {
			return nil
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			warnings = append(warnings, unreadablePathWarning(crawlRoot, filePath, err))
			return nil // zeroed metrics for this file
		}
		dir := dirRelativeToRoot(crawlRoot, filePath)
		// Every scanned directory appears in the map, zeros included —
		// the same contract as pythonAbstractnessScan.
		if _, ok := results[dir]; !ok {
			results[dir] = 0
		}
		var classes []cohesionClass
		switch language {
		case languagePython:
			classes = scanPythonCohesion(string(data))
		case languageTypeScript:
			classes = scanTypeScriptCohesion(string(data))
		}
		for _, class := range classes {
			results[dir] += methodComponents(class.methods)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return results, warnings, nil
}

// scanPythonCohesion extracts class declarations and their bound methods
// from Python source using an indentation stack. A def belongs to the
// innermost class frame directly below it in the stack (a def directly
// inside a class body); defs nested inside other defs, or at module
// level, are not method nodes. Instance-field references (self.X) are
// collected from every line of a method's body, including lines inside
// nested defs, and attributed to the nearest enclosing bound method.
func scanPythonCohesion(source string) []cohesionClass {
	type frame struct {
		isClass     bool
		indent      int
		classIndex  int
		methodIndex int
		bound       bool
	}
	var frames []frame
	var classes []cohesionClass

	for _, raw := range strings.Split(source, "\n") {
		trimmed := strings.TrimLeft(raw, " \t")
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := lineIndent(raw)

		for len(frames) > 0 && frames[len(frames)-1].indent >= indent {
			frames = frames[:len(frames)-1]
		}

		if match := pyClassRegex.FindStringSubmatch(trimmed); match != nil {
			classes = append(classes, cohesionClass{name: match[1]})
			frames = append(frames, frame{isClass: true, indent: indent, classIndex: len(classes) - 1})
			continue
		}
		if match := pyDefRegex.FindStringSubmatch(trimmed); match != nil {
			bound := false
			classIndex := -1
			methodIndex := -1
			if len(frames) > 0 && frames[len(frames)-1].isClass {
				bound = true
				classIndex = frames[len(frames)-1].classIndex
				classes[classIndex].methods = append(classes[classIndex].methods, cohesionMethod{
					name:   match[2],
					fields: make(map[string]bool),
				})
				methodIndex = len(classes[classIndex].methods) - 1
			}
			frames = append(frames, frame{
				indent:      indent,
				classIndex:  classIndex,
				methodIndex: methodIndex,
				bound:       bound,
			})
			continue
		}

		// Plain line: collect instance-field references for the nearest
		// enclosing bound method, if any.
		for i := len(frames) - 1; i >= 0; i-- {
			if frames[i].bound && !frames[i].isClass {
				method := &classes[frames[i].classIndex].methods[frames[i].methodIndex]
				for _, field := range pySelfFieldRegex.FindAllStringSubmatch(trimmed, -1) {
					method.fields[field[1]] = true
				}
				break
			}
		}
	}
	return classes
}

// scanTypeScriptCohesion extracts class declarations and their method
// members from TypeScript source using brace-depth tracking. A method is
// a same-line "name(params) {" declaration at the class body's depth;
// instance-field references (this.X) are collected from every deeper line
// until the method's brace closes.
func scanTypeScriptCohesion(source string) []cohesionClass {
	type frame struct {
		isClass     bool
		bodyDepth   int
		classIndex  int
		methodIndex int
		bound       bool
	}
	var frames []frame
	var classes []cohesionClass
	pendingClass := ""
	depth := 0

	for _, raw := range strings.Split(source, "\n") {
		// Frames are active while the current depth is inside them.
		for len(frames) > 0 && frames[len(frames)-1].bodyDepth > depth {
			frames = frames[:len(frames)-1]
		}

		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") {
			depth += strings.Count(raw, "{") - strings.Count(raw, "}")
			continue
		}

		if match := tsClassRegex.FindStringSubmatch(trimmed); match != nil {
			pendingClass = match[1]
		}

		if pendingClass != "" && strings.Contains(raw, "{") {
			classes = append(classes, cohesionClass{name: pendingClass})
			frames = append(frames, frame{
				isClass:    true,
				bodyDepth:  depth + 1,
				classIndex: len(classes) - 1,
			})
			pendingClass = ""
			depth += strings.Count(raw, "{") - strings.Count(raw, "}")
			continue
		}

		if len(frames) > 0 && frames[len(frames)-1].isClass && depth == frames[len(frames)-1].bodyDepth {
			if match := tsMethodRegex.FindStringSubmatch(raw); match != nil {
				classIndex := frames[len(frames)-1].classIndex
				classes[classIndex].methods = append(classes[classIndex].methods, cohesionMethod{
					name:   match[1],
					fields: make(map[string]bool),
				})
				frames = append(frames, frame{
					bodyDepth:   depth + 1,
					classIndex:  classIndex,
					methodIndex: len(classes[classIndex].methods) - 1,
					bound:       true,
				})
				depth += strings.Count(raw, "{") - strings.Count(raw, "}")
				continue
			}
		}

		for i := len(frames) - 1; i >= 0; i-- {
			if frames[i].bound && !frames[i].isClass {
				method := &classes[frames[i].classIndex].methods[frames[i].methodIndex]
				for _, field := range tsThisFieldRegex.FindAllStringSubmatch(trimmed, -1) {
					method.fields[field[1]] = true
				}
				break
			}
		}

		depth += strings.Count(raw, "{") - strings.Count(raw, "}")
	}
	return classes
}

// methodComponents returns the number of connected components in the
// method graph: methods are nodes, and two methods are connected when
// they share at least one instance-field reference (LCOM4, Hitz &
// Montazeri 1995). A method touching no fields is its own component; an
// empty method list has zero components. The count is invariant to
// union order, so the result is deterministic.
func methodComponents(methods []cohesionMethod) int {
	if len(methods) == 0 {
		return 0
	}
	parent := make([]int, len(methods))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		for parent[x] != x {
			x = parent[x]
		}
		return x
	}
	// Union every method with the first method seen touching each field.
	fieldOwner := make(map[string]int)
	for i, method := range methods {
		for field := range method.fields {
			if owner, ok := fieldOwner[field]; ok {
				ri, rj := find(i), find(owner)
				if ri != rj {
					parent[ri] = rj
				}
			} else {
				fieldOwner[field] = i
			}
		}
	}
	roots := make(map[int]bool, len(methods))
	for i := range methods {
		roots[find(i)] = true
	}
	return len(roots)
}

// lineIndent measures a line's leading whitespace: one per space, four
// per tab (the light-parse normalization).
func lineIndent(line string) int {
	n := 0
	for _, r := range line {
		if r == ' ' {
			n++
		} else if r == '\t' {
			n += 4
		} else {
			break
		}
	}
	return n
}
