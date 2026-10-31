package ripwireadapter

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// LCOM4 light-parse tests — openspec/changes/ripwire-adapter task 5.2.
// They pin the golden fixture values (fixturepy LCOM 6, src LCOM 2) and
// the union-find component counting on synthetic shapes.

func TestLCOM4Scan_FixturePython(t *testing.T) {
	t.Parallel()

	results, warnings, err := lcom4Scan(filepath.Join("testdata", "python", "src"), languagePython)
	if err != nil {
		t.Fatalf("lcom4Scan: %v", err)
	}

	// fixturepy = Cohesive(1) + SplitField(2) + AbstractThing(1) +
	// Concrete(1) + _Internal(1) = 6. Underscore classes count: the
	// spec's node definition is deliberately broader than the Go
	// adapter's exported-only filter.
	want := map[string]int{
		"fixturepy":            6,
		"fixturepy/cycle_a":    0,
		"fixturepy/cycle_b":    0,
		"fixturepy/operations": 0,
	}
	if !reflect.DeepEqual(results, want) {
		t.Errorf("results: got %v, want %v", results, want)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings: got %v, want none for a fully readable fixture", warnings)
	}
}

func TestLCOM4Scan_FixtureTypeScript(t *testing.T) {
	t.Parallel()

	results, warnings, err := lcom4Scan(filepath.Join("testdata", "typescript"), languageTypeScript)
	if err != nil {
		t.Fatalf("lcom4Scan: %v", err)
	}

	want := map[string]int{
		"src":         2, // SplitMethods: two field groups
		"src/cycle-a": 0,
		"src/cycle-b": 0,
	}
	if !reflect.DeepEqual(results, want) {
		t.Errorf("results: got %v, want %v", results, want)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings: got %v, want none for a fully readable fixture", warnings)
	}
}

func TestScanPythonCohesion_Synthetic(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		source    string
		wantTotal int // sum of per-class components
	}{
		{
			name: "module function excluded",
			source: "def module_function(value):\n" +
				"    return value * 2\n",
			wantTotal: 0,
		},
		{
			name: "class with no methods",
			source: "class Empty:\n" +
				"    pass\n",
			wantTotal: 0,
		},
		{
			name: "single method single component",
			source: "class A:\n" +
				"    def m(self):\n" +
				"        return \"x\"\n",
			wantTotal: 1,
		},
		{
			name: "methods sharing one field are one component",
			source: "class A:\n" +
				"    def w(self, v):\n" +
				"        self.f = v\n" +
				"    def r(self):\n" +
				"        return self.f\n",
			wantTotal: 1,
		},
		{
			name: "disjoint field groups are two components",
			source: "class A:\n" +
				"    def wx(self, v):\n" +
				"        self.x = v\n" +
				"    def wy(self, v):\n" +
				"        self.y = v\n",
			wantTotal: 2,
		},
		{
			name: "constructor bridges groups",
			source: "class A:\n" +
				"    def __init__(self):\n" +
				"        self.x = 0\n" +
				"        self.y = 0\n" +
				"    def wx(self, v):\n" +
				"        self.x = v\n" +
				"    def wy(self, v):\n" +
				"        self.y = v\n",
			wantTotal: 1,
		},
		{
			name: "nested def attributes to enclosing method",
			source: "class A:\n" +
				"    def outer(self):\n" +
				"        def inner():\n" +
				"            return self.z\n" +
				"        return inner\n" +
				"    def other(self):\n" +
				"        return self.z\n",
			wantTotal: 1,
		},
		{
			name: "two classes summed",
			source: "class A:\n" +
				"    def m(self):\n" +
				"        return self.p\n" +
				"class B:\n" +
				"    def m(self):\n" +
				"        return self.p\n" +
				"    def n(self):\n" +
				"        return self.q\n",
			wantTotal: 3,
		},
		{
			name: "async def is a method",
			source: "class A:\n" +
				"    async def m(self):\n" +
				"        return self.f\n",
			wantTotal: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			classes := scanPythonCohesion(tc.source)
			total := 0
			for _, class := range classes {
				total += methodComponents(class.methods)
			}
			if total != tc.wantTotal {
				t.Errorf("total components: got %d, want %d (classes %+v)", total, tc.wantTotal, classes)
			}
		})
	}
}

func TestScanTypeScriptCohesion_Synthetic(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		source    string
		wantTotal int
	}{
		{
			name:      "no classes",
			source:    "export interface Point {\n  x: number;\n}\n",
			wantTotal: 0,
		},
		{
			name: "single method",
			source: "class A {\n" +
				"  m(): void {\n" +
				"    return;\n" +
				"  }\n" +
				"}\n",
			wantTotal: 1,
		},
		{
			name: "shared field one component",
			source: "class A {\n" +
				"  w(v: number): void {\n" +
				"    this.f = v;\n" +
				"  }\n" +
				"  r(): number {\n" +
				"    return this.f;\n" +
				"  }\n" +
				"}\n",
			wantTotal: 1,
		},
		{
			name: "property declarations are not methods and do not bridge",
			source: "class A {\n" +
				"  x = 0;\n" +
				"  y = 0;\n" +
				"  setX(v: number): void {\n" +
				"    this.x = v;\n" +
				"  }\n" +
				"  setY(v: number): void {\n" +
				"    this.y = v;\n" +
				"  }\n" +
				"}\n",
			wantTotal: 2,
		},
		{
			name: "modifiers and return types",
			source: "class A {\n" +
				"  private async load(): Promise<void> {\n" +
				"    this.data = 1;\n" +
				"  }\n" +
				"  public read(): number {\n" +
				"    return this.data;\n" +
				"  }\n" +
				"}\n",
			wantTotal: 1,
		},
		{
			name: "constructor counts and bridges",
			source: "class A {\n" +
				"  constructor() {\n" +
				"    this.x = 0;\n" +
				"    this.y = 0;\n" +
				"  }\n" +
				"  setX(v: number): void {\n" +
				"    this.x = v;\n" +
				"  }\n" +
				"  setY(v: number): void {\n" +
				"    this.y = v;\n" +
				"  }\n" +
				"}\n",
			wantTotal: 1,
		},
		{
			name: "namespace interfaces are not classes",
			source: "namespace Shapes {\n" +
				"  export interface Point {\n" +
				"    x: number;\n" +
				"  }\n" +
				"}\n",
			wantTotal: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			classes := scanTypeScriptCohesion(tc.source)
			total := 0
			for _, class := range classes {
				total += methodComponents(class.methods)
			}
			if total != tc.wantTotal {
				t.Errorf("total components: got %d, want %d (classes %+v)", total, tc.wantTotal, classes)
			}
		})
	}
}

func TestMethodComponents(t *testing.T) {
	t.Parallel()

	mk := func(fields ...string) cohesionMethod {
		m := cohesionMethod{name: "m", fields: make(map[string]bool)}
		for _, f := range fields {
			m.fields[f] = true
		}
		return m
	}

	cases := []struct {
		name    string
		methods []cohesionMethod
		want    int
	}{
		{name: "empty", methods: nil, want: 0},
		{name: "one fieldless method", methods: []cohesionMethod{mk()}, want: 1},
		{name: "two isolated", methods: []cohesionMethod{mk("a"), mk("b")}, want: 2},
		{name: "two sharing", methods: []cohesionMethod{mk("a"), mk("a")}, want: 1},
		{
			name:    "chain",
			methods: []cohesionMethod{mk("a"), mk("a", "b"), mk("b"), mk("c")},
			want:    2,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := methodComponents(tc.methods); got != tc.want {
				t.Errorf("methodComponents: got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestLineIndent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		line string
		want int
	}{
		{name: "empty line", line: "", want: 0},
		{name: "four leading spaces", line: "    return x", want: 4},
		{name: "eight leading spaces", line: "        return x", want: 8},
		{name: "one leading tab", line: "\treturn x", want: 4},
		{name: "two leading tabs", line: "\t\treturn x", want: 8},
		{name: "two spaces then tab", line: "  \treturn x", want: 6},
		{name: "tab then two spaces", line: "\t  return x", want: 6},
		{name: "tab-only line", line: "\t", want: 4},
		{name: "two tabs only", line: "\t\t", want: 8},
		{name: "spaces-only line", line: "      ", want: 6},
		{name: "first rune non-whitespace", line: "x", want: 0},
		{name: "interior whitespace after non-whitespace not counted", line: "  x\t\t y", want: 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := lineIndent(tc.line); got != tc.want {
				t.Errorf("lineIndent(%q): got %d, want %d", tc.line, got, tc.want)
			}
		})
	}
}

func TestLCOM4Scan_Errors(t *testing.T) {
	t.Parallel()

	if _, _, err := lcom4Scan(filepath.Join(t.TempDir(), "absent"), languagePython); err == nil {
		t.Error("got no error, want error for missing crawl root")
	}
	if _, _, err := lcom4Scan(t.TempDir(), "cobol"); err == nil {
		t.Error("got no error, want error for unsupported language")
	}

	file := filepath.Join(t.TempDir(), "x.py")
	if err := os.WriteFile(file, []byte("class X:\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := lcom4Scan(file, languagePython); err == nil {
		t.Error("got no error, want error for non-directory crawl root")
	}
}

func TestLCOM4Scan_Deterministic(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		root, language string
	}{
		{filepath.Join("testdata", "python", "src"), languagePython},
		{filepath.Join("testdata", "typescript"), languageTypeScript},
	} {
		first, _, err := lcom4Scan(tc.root, tc.language)
		if err != nil {
			t.Fatalf("lcom4Scan: %v", err)
		}
		second, _, err := lcom4Scan(tc.root, tc.language)
		if err != nil {
			t.Fatalf("lcom4Scan: %v", err)
		}
		if !reflect.DeepEqual(first, second) {
			t.Errorf("%s: scanning twice yielded different results", tc.language)
		}
	}
}
