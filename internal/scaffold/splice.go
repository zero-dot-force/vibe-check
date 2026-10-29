package scaffold

import (
	"fmt"
	"sort"
	"strings"
)

// This file holds the pure, byte-preserving content splices for the UF
// policy files (design D3/D4/D10): given a policy file's bytes, each splice
// returns the rewritten bytes with the canonical entry or item inserted in
// sorted position, or a notice explaining why nothing was edited. The
// splices never touch the filesystem; the guarded orchestration lives in
// register.go.

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
// an advisory notice. A reviewers: key carrying an inline value (such as
// "reviewers: []"), a content-less file, or a file without a reviewers:
// section produces a notice and no edit. Entries are recognized in
// block-scalar style only; flow-collection entries (such as
// "- {agent: …}") are outside the presence check's contract.
func spliceCapabilities(content []byte, reg registration) spliceResult {
	lines := strings.Split(string(content), "\n")
	result := spliceResult{content: content}

	reviewersIdx := findTopLevelKey(lines, "reviewers")
	if reviewersIdx == -1 {
		result.notices = append(result.notices, fmt.Sprintf(
			"%s: no top-level reviewers: section; entry not inserted", reg.capabilitiesPath))
		return result
	}
	if inline := keyInlineValue(lines[reviewersIdx]); inline != "" {
		result.notices = append(result.notices, fmt.Sprintf(
			"%s: reviewers: section carries inline value %q; entry not inserted", reg.capabilitiesPath, inline))
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
	// entry's block (backed up past any trailing comments); else
	// immediately after the reviewers: line.
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
		insertIdx = substantiveEnd(lines, anchor.end)
		indent = leadingWhitespace(lines[anchor.start])
		for _, b := range reviewBlocks {
			if b.name > reg.agent {
				insertIdx = b.start
				indent = leadingWhitespace(lines[b.start])
				break
			}
		}
	}

	entry := capabilitiesEntryLines(reg.agent, indent)
	lines = insertLines(lines, entry, clampInsertIndex(lines, insertIdx))
	result.content = []byte(strings.Join(lines, "\n"))
	result.inserted = true
	return result
}

// spliceMatrix returns the review matrix content with the reg.agent item
// spliced into each advisor mode list per design D3/D4. Each mode list is
// an independent registration unit: a list already containing the item is
// left byte-identical, a missing, inline-valued ("code: []"), or item-less
// list emits a notice, and only existing item lists gain the item. The
// result is reported present when at least one mode list exists and every
// existing mode list already contains the item. Items are recognized in
// block-scalar style only; flow-collection items (such as
// "- {name: …}") are outside the presence check's contract.
func spliceMatrix(content []byte, reg registration) spliceResult {
	lines := strings.Split(string(content), "\n")
	result := spliceResult{content: content}

	advisorIdx := findTopLevelKey(lines, "advisor")
	if advisorIdx == -1 {
		result.notices = append(result.notices, fmt.Sprintf(
			"%s: no top-level advisor: section; item not inserted", reg.matrixPath))
		return result
	}
	if inline := keyInlineValue(lines[advisorIdx]); inline != "" {
		result.notices = append(result.notices, fmt.Sprintf(
			"%s: advisor: section carries inline value %q; item not inserted", reg.matrixPath, inline))
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
		if inline := keyInlineValue(lines[modeIdx]); inline != "" {
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

	// Apply the splices in descending file-position order so each insertion
	// leaves the remaining (lower) indices valid, whatever physical order
	// the mode lists appear in; each list is spliced independently
	// (design D4).
	sort.Slice(splices, func(i, j int) bool { return splices[i].insertIdx > splices[j].insertIdx })
	for _, s := range splices {
		lines = insertLines(lines, []string{s.line}, clampInsertIndex(lines, s.insertIdx))
	}
	if len(splices) > 0 {
		result.content = []byte(strings.Join(lines, "\n"))
		result.inserted = true
	}
	result.present = lists > 0 && presentLists == lists
	return result
}

// capabilitiesEntryLines returns the canonical reviewers: entry for agent,
// every line prefixed with indent plus the entry's own relative layout
// (design D2): the "- agent:" marker sits at indent, the capability and
// scopes keys two columns deeper, and the scope item four columns deeper.
func capabilitiesEntryLines(agent, indent string) []string {
	return []string{
		indent + "- agent: " + agent,
		indent + "  capability: " + reviewCapability,
		indent + "  scopes:",
		indent + "    - standard",
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
// whichever comes first. Entries are recognized in block-scalar style only
// (design D3); a flow-collection entry line is not a "- agent:" marker.
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

// substantiveEnd backs end up past any trailing blank or comment-only
// lines, so an insertion anchored at the end of the last entry lands after
// the entry's last substantive line rather than beyond trailing comments
// that belong to the section.
func substantiveEnd(lines []string, end int) int {
	for end > 0 && strings.TrimSpace(stripComment(lines[end-1])) == "" {
		end--
	}
	return end
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
