package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// This file implements the registration pass from the
// init-registers-divisor-entropy change (design D1-D10): after deploying
// assets, init splices the registered review agent into the target
// project's Unbound Force policy files. The pass is line-based (no YAML
// parser) so hand-curated comments and formatting survive verbatim, and it
// adds no package imports beyond those scaffold already depends on.

const (
	// registeredAgent is the review agent that init registers in a target
	// project's Unbound Force policy files.
	registeredAgent = "divisor-entropy"

	// capabilitiesPath is the UF policy file holding the reviewers:
	// section, relative to the target root.
	capabilitiesPath = ".uf/reviewer-capabilities.yaml"

	// matrixPath is the UF policy file holding the advisor: section,
	// relative to the target root.
	matrixPath = ".uf/review-matrix.yaml"

	// capabilitiesMarkerFallback is the "- " marker indentation, in spaces,
	// used for an inserted reviewers: entry when no sibling entry supplies
	// one. It matches the UF scaffold default.
	capabilitiesMarkerFallback = 2

	// matrixItemFallback is the item indentation, in spaces, used for an
	// inserted advisor list item when no sibling item supplies one. It
	// matches the UF scaffold default.
	matrixItemFallback = 4

	// reviewCapability is the capability a reviewers: entry must declare to
	// satisfy review dispatch.
	reviewCapability = "review"
)

// registration describes one review agent that init registers into a target
// project's Unbound Force policy files (design D2). The one-entry table
// models the concept "review personas this tool deploys" so a future review
// agent is a localized addition rather than speculative generality.
type registration struct {
	// agent is the canonical agent name inserted into both policy files.
	agent string
	// capabilitiesPath is the policy file holding the reviewers: section,
	// relative to the target root.
	capabilitiesPath string
	// matrixPath is the policy file holding the advisor: section, relative
	// to the target root.
	matrixPath string
	// matrixModes are the advisor mode lists that each gain the agent item.
	matrixModes []string
}

// registrations drives the registration pass; exactly one agent is
// registered today.
var registrations = []registration{
	{
		agent:            registeredAgent,
		capabilitiesPath: capabilitiesPath,
		matrixPath:       matrixPath,
		matrixModes:      []string{"code", "spec", "triage", "feedback"},
	},
}

// spliceResult reports the outcome of registering one agent in one policy
// file's content.
type spliceResult struct {
	// content is the rewritten file content; byte-identical to the input
	// when nothing was inserted.
	content []byte
	// inserted is true when at least one registration unit gained the
	// canonical entry.
	inserted bool
	// present is true when the file is present in a recognized shape and
	// every registration unit already contains the entry (design D4).
	present bool
	// notices describes unexpected shapes and advisory gaps; each notice
	// names the policy file it concerns.
	notices []string
}

// spliceCapabilities returns the capabilities manifest content with the
// canonical entry for reg.agent spliced in per design D3/D10. Scanning is
// confined to the top-level reviewers: section; a present entry suppresses
// insertion regardless of its capability or scopes (value-blind), and a
// present entry lacking capability: review or declaring empty scopes emits
// an advisory notice. Content without a reviewers: section produces a
// notice and no edit.
func spliceCapabilities(content []byte, reg registration) spliceResult {
	lines := strings.Split(string(content), "\n")
	result := spliceResult{content: content}

	reviewersIdx := findTopLevelKey(lines, "reviewers")
	if reviewersIdx == -1 {
		result.notices = append(result.notices, fmt.Sprintf(
			"%s: no top-level reviewers: section; entry not inserted", reg.capabilitiesPath))
		return result
	}
	sectionEnd := nextTopLevelKey(lines, reviewersIdx+1)

	blocks := entryBlocks(lines, reviewersIdx+1, sectionEnd)

	// Presence is value-blind and scope-confined (design D10): any entry in
	// the section named reg.agent suppresses insertion.
	for _, b := range blocks {
		if b.name != reg.agent {
			continue
		}
		result.present = true
		if notice, ok := presenceAdvisory(lines, b, reg); ok {
			result.notices = append(result.notices, notice)
		}
		return result
	}

	// Sorted insertion (design D3): before the first capability: review
	// entry whose name sorts after the agent; else after the last review
	// entry's block; else immediately after the reviewers: line.
	insertIdx := reviewersIdx + 1
	indent := strings.Repeat(" ", capabilitiesMarkerFallback)
	var reviewBlocks []entryBlock
	for _, b := range blocks {
		if entryDeclares(lines, b, "capability", reviewCapability) {
			reviewBlocks = append(reviewBlocks, b)
		}
	}
	if len(reviewBlocks) > 0 {
		anchor := reviewBlocks[len(reviewBlocks)-1]
		insertIdx = anchor.end
		indent = leadingWhitespace(lines[anchor.start])
		for _, b := range reviewBlocks {
			if b.name > reg.agent {
				insertIdx = b.start
				indent = leadingWhitespace(lines[b.start])
				break
			}
		}
	}

	entry := make([]string, 0, len(capabilitiesEntryLines(reg.agent)))
	for i, line := range capabilitiesEntryLines(reg.agent) {
		prefix := indent
		switch i {
		case 1, 2:
			prefix += "  "
		case 3:
			prefix += "    "
		}
		entry = append(entry, prefix+line)
	}

	lines = insertLines(lines, entry, clampInsertIndex(lines, insertIdx))
	result.content = []byte(strings.Join(lines, "\n"))
	result.inserted = true
	return result
}

// spliceMatrix returns the review matrix content with the reg.agent item
// spliced into each advisor mode list per design D3/D4. Each mode list is an
// independent registration unit: a list already containing the item is left
// byte-identical, a missing or unrecognizable list emits a notice, and only
// existing item lists gain the item. The result is reported present when at
// least one mode list exists and every existing mode list already contains
// the item.
func spliceMatrix(content []byte, reg registration) spliceResult {
	lines := strings.Split(string(content), "\n")
	result := spliceResult{content: content}

	advisorIdx := findTopLevelKey(lines, "advisor")
	if advisorIdx == -1 {
		result.notices = append(result.notices, fmt.Sprintf(
			"%s: no top-level advisor: section; item not inserted", reg.matrixPath))
		return result
	}
	sectionEnd := nextTopLevelKey(lines, advisorIdx+1)

	type listSplice struct {
		insertIdx int
		line      string
	}
	var splices []listSplice
	lists, presentLists := 0, 0

	for _, mode := range reg.matrixModes {
		modeIdx := findModeKey(lines, mode, advisorIdx+1, sectionEnd)
		if modeIdx == -1 {
			result.notices = append(result.notices, fmt.Sprintf(
				"%s: advisor mode list %q absent; not created", reg.matrixPath, mode))
			continue
		}
		if inline := modeInlineValue(lines[modeIdx]); inline != "" {
			result.notices = append(result.notices, fmt.Sprintf(
				"%s: advisor mode list %q carries inline value %q; not edited",
				reg.matrixPath, mode, inline))
			continue
		}
		items := listItemIndices(lines, modeIdx+1, sectionEnd)
		if len(items) == 0 {
			result.notices = append(result.notices, fmt.Sprintf(
				"%s: advisor mode list %q has no items; not edited", reg.matrixPath, mode))
			continue
		}
		lists++

		present := false
		for _, idx := range items {
			value, _ := listValue(lines[idx])
			if extractName(value) == reg.agent {
				present = true
				break
			}
		}
		if present {
			presentLists++
			continue
		}

		// Sorted insertion (design D3): before the first item whose name
		// sorts after the agent, else after the last item, at the
		// indentation of the neighboring items.
		insertIdx := items[len(items)-1] + 1
		indent := leadingWhitespace(lines[items[len(items)-1]])
		for _, idx := range items {
			value, _ := listValue(lines[idx])
			if extractName(value) > reg.agent {
				insertIdx = idx
				indent = leadingWhitespace(lines[idx])
				break
			}
		}
		if indent == "" {
			indent = strings.Repeat(" ", matrixItemFallback)
		}
		splices = append(splices, listSplice{insertIdx: insertIdx, line: indent + matrixItemLine(reg.agent)})
	}

	// Apply splices bottom-up so earlier indices stay valid; each list is
	// spliced independently (design D4).
	for i := len(splices) - 1; i >= 0; i-- {
		lines = insertLines(lines, []string{splices[i].line}, clampInsertIndex(lines, splices[i].insertIdx))
	}
	if len(splices) > 0 {
		result.content = []byte(strings.Join(lines, "\n"))
		result.inserted = true
	}
	result.present = lists > 0 && presentLists == lists
	return result
}

// capabilitiesEntryLines returns the canonical reviewers: entry for agent
// without indentation (design D2): the marker line sits at the sibling
// entry indentation, the capability and scopes keys two columns deeper, and
// the scope item four columns deeper.
func capabilitiesEntryLines(agent string) []string {
	return []string{
		"- agent: " + agent,
		"capability: review",
		"scopes:",
		"- standard",
	}
}

// matrixItemLine returns the canonical advisor list item for agent without
// indentation (design D2).
func matrixItemLine(agent string) string {
	return "- " + agent
}

// entryBlock is one reviewers: entry: the "- agent:" marker line at start,
// the block running to end (exclusive), and the extracted agent name.
type entryBlock struct {
	start int
	end   int
	name  string
}

// entryBlocks parses the "- agent:" entry blocks within lines [from, end).
// Each block runs from its "- agent:" line to the next entry or to end,
// whichever comes first.
func entryBlocks(lines []string, from, end int) []entryBlock {
	var blocks []entryBlock
	for i := from; i < end; i++ {
		value, ok := listValue(lines[i])
		if !ok {
			continue
		}
		rest, found := strings.CutPrefix(value, "agent:")
		if !found {
			continue
		}
		if len(blocks) > 0 {
			blocks[len(blocks)-1].end = i
		}
		blocks = append(blocks, entryBlock{start: i, end: end, name: extractName(rest)})
	}
	return blocks
}

// presenceAdvisory inspects a present reviewers: entry and returns an
// advisory notice when the entry lacks capability: review or declares empty
// scopes (design D10). The entry itself is never rewritten.
func presenceAdvisory(lines []string, b entryBlock, reg registration) (string, bool) {
	capability, hasCapability := "", false
	hasScopes, emptyScopes := false, true
	for i := b.start + 1; i < b.end; i++ {
		if v, ok := mappingValue(lines[i], "capability"); ok {
			capability, hasCapability = v, true
		}
		if v, ok := mappingValue(lines[i], "scopes"); ok {
			hasScopes = true
			if v != "" && v != "[]" {
				emptyScopes = false
			}
			for j := i + 1; j < b.end; j++ {
				if _, ok := listValue(lines[j]); ok {
					emptyScopes = false
					break
				}
			}
		}
	}
	switch {
	case !hasCapability || capability != reviewCapability:
		shown := capability
		if !hasCapability {
			shown = "none"
		}
		return fmt.Sprintf("%s: %s present with capability: %s; review dispatch may fail closed",
			reg.capabilitiesPath, reg.agent, shown), true
	case !hasScopes || emptyScopes:
		return fmt.Sprintf("%s: %s present with empty scopes; review dispatch may fail closed",
			reg.capabilitiesPath, reg.agent), true
	}
	return "", false
}

// entryDeclares reports whether the entry block spanning lines [b.start,
// b.end) carries the mapping key with the given value.
func entryDeclares(lines []string, b entryBlock, key, value string) bool {
	for i := b.start + 1; i < b.end; i++ {
		if v, ok := mappingValue(lines[i], key); ok && v == value {
			return true
		}
	}
	return false
}

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

// findModeKey returns the index of the indented mapping key naming mode
// within lines [from, end), or -1. The scan is confined to the advisor
// section so mode-named keys elsewhere (for example the defaults: section)
// are never matched.
func findModeKey(lines []string, mode string, from, end int) int {
	for i := from; i < end; i++ {
		trimmed := strings.TrimLeft(lines[i], " \t")
		if trimmed == lines[i] || trimmed == "" {
			continue // mode keys sit one level below advisor:
		}
		if stripComment(trimmed) == mode+":" {
			return i
		}
	}
	return -1
}

// modeInlineValue returns the comment-stripped inline value of a mode key
// line: "  code:" yields "" and "  code: []" yields "[]".
func modeInlineValue(line string) string {
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
// of surrounding single or double quotes is removed (design D3).
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

// registerReviewAgent registers reg's agent into the UF policy files under
// root after asset deployment (design D1, D5-D9). It returns the
// category-prefixed outcome names ("reviewers/<agent>" for the capabilities
// manifest and "advisors/<agent>" for the review matrix) for the Written
// and Skipped result slices, plus any notices. Absent files and unexpected
// shapes are notice-skips that never create structure; symlink refusals and
// I/O failures return wrapped errors so the command exits 2 with the target
// untouched (design D9).
func registerReviewAgent(root string, reg registration, writeFile func(string, []byte, fs.FileMode) error) (written, skipped, notices []string, err error) {
	capResult, err := registerPolicyFile(root, reg, reg.capabilitiesPath, spliceCapabilities, writeFile)
	if err != nil {
		return nil, nil, nil, err
	}
	matrixResult, err := registerPolicyFile(root, reg, reg.matrixPath, spliceMatrix, writeFile)
	if err != nil {
		return nil, nil, nil, err
	}

	if capResult.inserted {
		written = append(written, "reviewers/"+reg.agent)
	} else if capResult.present {
		skipped = append(skipped, "reviewers/"+reg.agent)
	}
	if matrixResult.inserted {
		written = append(written, "advisors/"+reg.agent)
	} else if matrixResult.present {
		skipped = append(skipped, "advisors/"+reg.agent)
	}
	notices = append(notices, capResult.notices...)
	notices = append(notices, matrixResult.notices...)
	return written, skipped, notices, nil
}

// registerPolicyFile guards and rewrites one policy file per design D6-D8.
// The parent directory is Lstatted without being created — an absent
// directory or file is a notice-skip — while symlinks, non-directory
// components, and non-regular files are refused with wrapped errors. The
// new content is computed fully in memory and applied with a single write
// through the injectable writeFile seam, passing the file's existing mode so
// it is preserved; a write failure states that the file may be partially
// written and restorable from version control.
func registerPolicyFile(root string, reg registration, relPath string, splice func([]byte, registration) spliceResult, writeFile func(string, []byte, fs.FileMode) error) (spliceResult, error) {
	result := spliceResult{}

	dir := filepath.Join(root, filepath.Dir(relPath))
	dirInfo, err := os.Lstat(dir)
	switch {
	case err == nil:
		if dirInfo.Mode()&fs.ModeSymlink != 0 {
			return result, fmt.Errorf("scaffold: refusing to follow symlink in policy path: %s", dir)
		}
		if !dirInfo.IsDir() {
			return result, fmt.Errorf("scaffold: policy path component is not a directory: %s", dir)
		}
	case os.IsNotExist(err):
		result.notices = append(result.notices, relPath+": not present; not created")
		return result, nil
	default:
		return result, fmt.Errorf("scaffold: inspect %q: %w", dir, err)
	}

	path := filepath.Join(root, relPath)
	info, err := os.Lstat(path)
	switch {
	case err == nil:
		if info.Mode()&fs.ModeSymlink != 0 {
			return result, fmt.Errorf("scaffold: refusing to overwrite symlink: %s", path)
		}
		if !info.Mode().IsRegular() {
			return result, fmt.Errorf("scaffold: refusing to write non-regular policy file: %s", path)
		}
	case os.IsNotExist(err):
		result.notices = append(result.notices, relPath+": not present; not created")
		return result, nil
	default:
		return result, fmt.Errorf("scaffold: inspect %q: %w", path, err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return result, fmt.Errorf("scaffold: read policy file %q: %w", path, err)
	}

	result = splice(content, reg)
	if !result.inserted {
		return result, nil
	}

	if err := writeFile(path, result.content, info.Mode().Perm()); err != nil {
		return result, fmt.Errorf("scaffold: write policy file %q (may be partially written; restore from version control): %w", path, err)
	}
	return result, nil
}
