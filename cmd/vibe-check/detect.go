package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// detectLanguage determines the language of the project at path (task
// 7.1, go-analyze D10's deferred refactor). It is a pure function of the
// filesystem state at path — no subprocesses, no network, no
// configuration — and its decision order is deterministic:
//
//  1. go.mod present → "go"
//  2. pyproject.toml, setup.py, or setup.cfg present → "python"
//  3. tsconfig.json present → "typescript"
//  4. otherwise, a file-extension census of the tree (recursive,
//     case-insensitive): more .ts/.tsx files than .py files →
//     "typescript"; more .py files → "python"
//  5. a tie (including zero-of-both) → "go", the pre-existing behavior
//     for markerless trees: the Go adapter's own analysis surfaces the
//     authoritative "no Go packages" error rather than a language-guess
//     failure.
//
// Markers are checked at the top of path only (a monorepo's nested
// manifests are not consulted); the census is recursive but bounded by
// the same os.ReadDir walk every other scanner uses.
func detectLanguage(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("detect language: stat %s: %w", path, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("detect language: %s: not a directory", path)
	}

	// Priority order, first match wins: a tree with both a go.mod and a
	// tsconfig.json (a Go tool wrapping a TS project) is a Go module.
	markers := []struct{ marker, language string }{
		{"go.mod", "go"},
		{"pyproject.toml", "python"},
		{"setup.py", "python"},
		{"setup.cfg", "python"},
		{"tsconfig.json", "typescript"},
	}
	for _, m := range markers {
		if _, err := os.Stat(filepath.Join(path, m.marker)); err == nil {
			return m.language, nil
		}
	}

	py, ts, err := censusExtensions(path)
	if err != nil {
		return "", err
	}
	if ts > py {
		return "typescript", nil
	}
	if py > ts {
		return "python", nil
	}
	return "go", nil
}

// censusExtensions recursively counts Python (.py) and TypeScript
// (.ts/.tsx) files under root. Unreadable directories are skipped: a
// language guess must not fail because one subtree is locked, and the
// marker checks above already decided every manifest-bearing project.
func censusExtensions(root string) (py, ts int, err error) {
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // skip unreadable subtrees
		}
		if d.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(d.Name())) {
		case ".py":
			py++
		case ".ts", ".tsx":
			ts++
		}
		return nil
	})
	if err != nil {
		return 0, 0, fmt.Errorf("detect language: walk %s: %w", root, err)
	}
	return py, ts, nil
}
