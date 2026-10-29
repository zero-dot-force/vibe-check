package scaffold

import "strings"

// This file holds the line-scanning helpers shared by the UF policy
// splices: a minimal, indentation-aware reader for the block-scalar YAML
// subset the policy files use. The scanner is deliberately line-based (no
// YAML parser) so hand-curated comments, quoting, and formatting survive
// verbatim; every helper is pure and never mutates its input.

// findTopLevelKey returns the index of the first line that is a top-level
// YAML mapping key with the given name, or -1 when none exists.
func findTopLevelKey(lines []string, key string) int {
	for i, line := range lines {
		if k, ok := isTopLevelKey(line); ok && k == key {
			return i
		}
	}
	return -1
}

// nextTopLevelKey returns the index of the first top-level YAML mapping key
// at or after from, or len(lines) when none follows.
func nextTopLevelKey(lines []string, from int) int {
	for i := from; i < len(lines); i++ {
		if _, ok := isTopLevelKey(lines[i]); ok {
			return i
		}
	}
	return len(lines)
}

// isTopLevelKey reports whether the line is a YAML mapping key at zero
// indentation (^[A-Za-z_][A-Za-z0-9_]*:), tolerating a trailing comment,
// and returns the key.
func isTopLevelKey(line string) (string, bool) {
	if line == "" || !isKeyStart(line[0]) {
		return "", false
	}
	i := 0
	for i < len(line) && line[i] != ':' {
		if !isKeyChar(line[i]) {
			return "", false
		}
		i++
	}
	if i == 0 || i >= len(line) {
		return "", false
	}
	return line[:i], true
}

// isKeyStart reports whether c may begin a YAML mapping key.
func isKeyStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isKeyChar reports whether c may appear inside a YAML mapping key.
func isKeyChar(c byte) bool {
	return isKeyStart(c) || (c >= '0' && c <= '9')
}

// findModeKey returns the index of the first indented mapping key naming
// mode within lines [from, end), or -1. The key may carry an inline value
// ("code: []") — the caller notices such shapes instead of editing them.
// The scan is confined to the advisor section so mode-named keys elsewhere
// (for example the defaults: section) are never matched.
func findModeKey(lines []string, mode string, from, end int) int {
	for i := from; i < end; i++ {
		trimmed := strings.TrimLeft(lines[i], " \t")
		if trimmed == lines[i] || trimmed == "" {
			continue // mode keys sit one level below advisor:
		}
		if strings.HasPrefix(stripComment(trimmed), mode+":") {
			return i
		}
	}
	return -1
}

// keyInlineValue returns the comment-stripped inline value of a YAML key
// line: "code:" yields "" and "code: []" (or "reviewers: {}") yields the
// flow value. A key whose value lives in block style below it yields "".
func keyInlineValue(line string) string {
	trimmed := strings.TrimSpace(line)
	i := strings.Index(trimmed, ":")
	if i < 0 {
		return ""
	}
	return stripComment(trimmed[i+1:])
}

// listItemIndices returns the indices of the sequence item lines forming
// the list starting at from: the run of "- …" items below a mode key, with
// blank lines and comments skipped over structurally (never rewritten). The
// run ends at the first line that is neither an item nor a blank line nor a
// comment.
func listItemIndices(lines []string, from, end int) []int {
	var items []int
	for i := from; i < end; i++ {
		line := lines[i]
		if strings.TrimSpace(stripComment(line)) == "" {
			continue
		}
		if _, ok := listValue(line); !ok {
			break
		}
		items = append(items, i)
	}
	return items
}

// listValue returns the scalar value of a YAML sequence item line
// ("- value") and reports whether the line is a sequence item at all.
func listValue(line string) (string, bool) {
	trimmed := strings.TrimLeft(line, " \t")
	rest, ok := strings.CutPrefix(trimmed, "-")
	if !ok {
		return "", false
	}
	if rest != "" && rest[0] != ' ' && rest[0] != '\t' {
		return "", false
	}
	return strings.TrimLeft(rest, " \t"), true
}

// mappingValue returns the normalized value of an indented mapping line
// carrying the given key ("  key: value"), reporting whether the line
// carries the key at all. The value is normalized by extractName; a key
// line with no value yields "".
func mappingValue(line string, key string) (string, bool) {
	trimmed := strings.TrimLeft(line, " \t")
	rest, ok := strings.CutPrefix(trimmed, key+":")
	if !ok {
		return "", false
	}
	return extractName(rest), true
}

// extractName normalizes a YAML scalar for comparison: surrounding
// whitespace is trimmed, a trailing "# …" comment is stripped, and one pair
// of surrounding single or double quotes is removed (design D3). Only
// block-scalar values are recognized; flow-collection values (such as
// "{agent: …}") are outside this contract and compare unequal to any
// block-scalar name.
func extractName(raw string) string {
	s := strings.TrimSpace(raw)
	if i := strings.Index(s, "#"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			s = s[1 : len(s)-1]
		}
	}
	return strings.TrimSpace(s)
}

// stripComment removes a trailing "# …" comment and surrounding whitespace.
func stripComment(s string) string {
	if i := strings.Index(s, "#"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

// leadingWhitespace returns the leading spaces and tabs of a line.
func leadingWhitespace(line string) string {
	return line[:len(line)-len(strings.TrimLeft(line, " \t"))]
}

// insertLines inserts the given lines before index i, returning a new
// slice. When i is the end of a file whose final line lacks a terminator,
// the caller's join appends the missing terminator so the inserted block
// starts on its own line (design D8).
func insertLines(lines []string, inserted []string, i int) []string {
	out := make([]string, 0, len(lines)+len(inserted))
	out = append(out, lines[:i]...)
	out = append(out, inserted...)
	out = append(out, lines[i:]...)
	return out
}

// clampInsertIndex adjusts an insertion index that points past the empty
// sentinel element strings.Split produces for a trailing terminator, so
// inserted content lands after the file's final real line rather than after
// the terminator itself.
func clampInsertIndex(lines []string, i int) int {
	if i > len(lines) {
		i = len(lines)
	}
	if i == len(lines) && len(lines) > 0 && lines[len(lines)-1] == "" {
		return len(lines) - 1
	}
	return i
}
