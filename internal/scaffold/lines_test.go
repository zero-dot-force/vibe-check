package scaffold

import "testing"

// This file holds table-driven unit tests for the pure line-scanning
// helpers in lines.go: the edge cases of the block-scalar YAML subset
// reader (top-level key recognition, inline values, sequence items, and
// insertion-index clamping).

func TestIsTopLevelKey(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		line string
		want string
		ok   bool
	}{
		{"key with value", "key: v", "key", true},
		{"key with empty value", "key:", "key", true},
		{"key with trailing comment", "key: v # comment", "key", true},
		{"empty line", "", "", false},
		{"indented key", "  indented: v", "", false},
		{"leading dash", "-foo: v", "", false},
		{"dash inside key", "key-x: v", "", false},
		{"no colon", "keyonly", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := isTopLevelKey(tc.line)
			if got != tc.want || ok != tc.ok {
				t.Errorf("isTopLevelKey(%q): got (%q, %v), want (%q, %v)", tc.line, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestKeyInlineValue(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		line string
		want string
	}{
		{"empty value", "code:", ""},
		{"flow value", "code: []", "[]"},
		{"comment-stripped value", "a: b # c", "b"},
		{"indented key", "  indented: value", "value"},
		{"no colon", "no colon", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := keyInlineValue(tc.line); got != tc.want {
				t.Errorf("keyInlineValue(%q): got %q, want %q", tc.line, got, tc.want)
			}
		})
	}
}

func TestListValue(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		line string
		want string
		ok   bool
	}{
		{"item with value", "- value", "value", true},
		{"item with tab separator", "-\tvalue", "value", true},
		{"indented item", "  - value", "value", true},
		{"item with empty value", "- ", "", true},
		{"bare dash", "-", "", true},
		{"dash fused to value", "-value", "", false},
		{"double dash", "--flag", "", false},
		{"plain value", "value", "", false},
		{"empty line", "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := listValue(tc.line)
			if got != tc.want || ok != tc.ok {
				t.Errorf("listValue(%q): got (%q, %v), want (%q, %v)", tc.line, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestClampInsertIndex(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		lines []string
		i     int
		want  int
	}{
		{"index past end clamped", []string{"a", "b"}, 5, 2},
		{"index at end", []string{"a", "b"}, 2, 2},
		{"index mid file", []string{"a", "b"}, 1, 1},
		{"trailing terminator sentinel", []string{"a", "b", ""}, 3, 2},
		{"past end with terminator", []string{"a", "b", ""}, 5, 2},
		{"empty lines zero index", nil, 0, 0},
		{"empty lines past end", nil, 2, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := clampInsertIndex(tc.lines, tc.i); got != tc.want {
				t.Errorf("clampInsertIndex(len=%d, i=%d): got %d, want %d", len(tc.lines), tc.i, got, tc.want)
			}
		})
	}
}
