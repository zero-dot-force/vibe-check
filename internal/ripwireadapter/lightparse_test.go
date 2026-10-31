package ripwireadapter

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Python abstractness light-parse tests — openspec/changes/ripwire-adapter
// task 5.1. They pin design D8: the abc.ABC class counts in both
// AbstractTypes and ExportedTypes, concrete classes only in ExportedTypes,
// underscore classes in neither, and pruned denylist directories nowhere.

func TestPythonAbstractnessScan_Fixture(t *testing.T) {
	t.Parallel()

	results, warnings, err := pythonAbstractnessScan(filepath.Join("testdata", "python", "src"))
	if err != nil {
		t.Fatalf("pythonAbstractnessScan: %v", err)
	}

	want := map[string]lightParseResult{
		"fixturepy":            {AbstractTypes: 1, ExportedTypes: 4},
		"fixturepy/cycle_a":    {},
		"fixturepy/cycle_b":    {},
		"fixturepy/operations": {},
	}
	if !reflect.DeepEqual(results, want) {
		t.Errorf("results: got %v, want %v", results, want)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings: got %v, want none for a fully readable fixture", warnings)
	}
}

func TestScanPythonClasses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		source     string
		wantCounts lightParseResult
	}{
		{
			name:       "abc dot ABC",
			source:     "class Foo(abc.ABC):\n    pass\n",
			wantCounts: lightParseResult{AbstractTypes: 1, ExportedTypes: 1},
		},
		{
			name:       "bare ABC import form",
			source:     "class Foo(ABC):\n    pass\n",
			wantCounts: lightParseResult{AbstractTypes: 1, ExportedTypes: 1},
		},
		{
			name:       "typing Protocol",
			source:     "class Foo(typing.Protocol):\n    pass\n",
			wantCounts: lightParseResult{AbstractTypes: 1, ExportedTypes: 1},
		},
		{
			name:       "generic protocol",
			source:     "class Foo(Protocol[T], Base):\n    pass\n",
			wantCounts: lightParseResult{AbstractTypes: 1, ExportedTypes: 1},
		},
		{
			name:       "dotted ABCMeta alias",
			source:     "class Foo(meta.ABCMeta, Base):\n    pass\n",
			wantCounts: lightParseResult{AbstractTypes: 1, ExportedTypes: 1},
		},
		{
			name:       "concrete only exported",
			source:     "class Foo(BaseClass):\n    pass\n",
			wantCounts: lightParseResult{AbstractTypes: 0, ExportedTypes: 1},
		},
		{
			name:       "no bases",
			source:     "class Foo:\n    pass\n",
			wantCounts: lightParseResult{AbstractTypes: 0, ExportedTypes: 1},
		},
		{
			name:       "underscore abstract not exported",
			source:     "class _Hidden(abc.ABC):\n    pass\n",
			wantCounts: lightParseResult{AbstractTypes: 1, ExportedTypes: 0},
		},
		{
			name:       "underscore concrete neither",
			source:     "class _Hidden:\n    pass\n",
			wantCounts: lightParseResult{AbstractTypes: 0, ExportedTypes: 0},
		},
		{
			name:       "metaclass keyword is not a base",
			source:     "class Foo(Base, metaclass=ABCMeta):\n    pass\n",
			wantCounts: lightParseResult{AbstractTypes: 0, ExportedTypes: 1},
		},
		{
			name:       "commented class ignored",
			source:     "# class Foo(abc.ABC):\n    pass\n",
			wantCounts: lightParseResult{},
		},
		{
			name:       "calls that merely start with class",
			source:     "classify(this)\nclassy = 1\n",
			wantCounts: lightParseResult{},
		},
		{
			name:       "empty source",
			source:     "",
			wantCounts: lightParseResult{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got lightParseResult
			for _, class := range scanPythonClasses(tc.source) {
				if class.abstract {
					got.AbstractTypes++
				}
				if class.exported {
					got.ExportedTypes++
				}
			}
			if got != tc.wantCounts {
				t.Errorf("scanPythonClasses: got %+v, want %+v", got, tc.wantCounts)
			}
		})
	}
}

func TestPythonAbstractnessScan_SyntheticTree(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	write := func(rel, content string) {
		t.Helper()
		full := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("pkg/mod.py", "class A(abc.ABC):\n    pass\n\nclass B:\n    pass\n\nclass _C(abc.ABC):\n    pass\n")
	write("pkg/build/gen.py", "class Pruned(abc.ABC):\n    pass\n")
	write("notes.txt", "class NotPython(abc.ABC):\n")

	results, warnings, err := pythonAbstractnessScan(root)
	if err != nil {
		t.Fatalf("pythonAbstractnessScan: %v", err)
	}

	want := map[string]lightParseResult{"pkg": {AbstractTypes: 2, ExportedTypes: 2}}
	if !reflect.DeepEqual(results, want) {
		t.Errorf("results: got %v, want %v", results, want)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings: got %v, want none", warnings)
	}
}

func TestPythonAbstractnessScan_Errors(t *testing.T) {
	t.Parallel()

	t.Run("missing root", func(t *testing.T) {
		if _, _, err := pythonAbstractnessScan(filepath.Join(t.TempDir(), "absent")); err == nil {
			t.Fatal("got no error, want error for missing crawl root")
		}
	})

	t.Run("root is a file", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "file.py")
		if err := os.WriteFile(file, []byte("class X:\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, _, err := pythonAbstractnessScan(file); err == nil {
			t.Fatal("got no error, want error for non-directory crawl root")
		}
	})
}
