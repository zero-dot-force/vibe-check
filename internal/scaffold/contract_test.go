package scaffold

import (
	"io/fs"
	"path"
	"slices"
	"strings"
	"testing"
)

// TestEmbeddedAsset_DivisorEntropyContract covers the embedded
// divisor-entropy.md: required frontmatter, provenance marker, sections, and
// the bash allowlist.
func TestEmbeddedAsset_DivisorEntropyContract(t *testing.T) {
	t.Parallel()
	data, err := fs.ReadFile(agentAssetsFS, embeddedAgentAssetPath)
	if err != nil {
		t.Fatalf("read embedded asset: %v", err)
	}
	content := string(data)

	if ok, _ := path.Match("divisor-*.md", deployedAgentAssetName); !ok {
		t.Errorf("asset name %q does not match divisor-*.md glob", deployedAgentAssetName)
	}
	matches, err := fs.Glob(agentAssetsFS, "assets/agents/divisor-*.md")
	if err != nil {
		t.Fatalf("glob embedded assets: %v", err)
	}
	if !slices.Contains(matches, embeddedAgentAssetPath) {
		t.Errorf("embedded assets missing %q; got %v", embeddedAgentAssetPath, matches)
	}

	for _, needle := range []string{
		"mode: subagent",
		"temperature: 0.1",
		"edit: deny",
		"webfetch: deny",
		`"*": "deny"`,
	} {
		if !strings.Contains(content, needle) {
			t.Errorf("frontmatter missing %q", needle)
		}
	}

	if frontmatterDescription(content) == "" {
		t.Errorf("frontmatter description is empty or missing")
	}

	if !strings.Contains(content, "<!-- scaffolded by vibe-check") {
		t.Errorf("missing AP-006 provenance marker prefix")
	}

	for _, section := range []string{
		"## Source Documents",
		"## Code Review Mode",
		"## Output Format",
		"## Decision Criteria",
		"## Security / Operating Constraints",
	} {
		if !strings.Contains(content, "\n"+section) {
			t.Errorf("missing required section heading %q", section)
		}
	}

	allow, denyCatchAll := bashPermissions(content)
	if !denyCatchAll {
		t.Errorf(`bash catch-all "*" must map to "deny"`)
	}
	want := []string{
		"git merge-base *",
		"git rev-parse *",
		"git worktree add *",
		"git worktree remove *",
		"git worktree prune",
		"git fetch origin *",
		"git check-ref-format *",
		"vibe-check analyze *",
		"vibe-check diff *",
	}
	if len(allow) != len(want) {
		t.Errorf("bash allow count: got %d %v, want %d", len(allow), allow, len(want))
	}
	wantSet := make(map[string]bool, len(want))
	for _, k := range want {
		wantSet[k] = true
		if !allow[k] {
			t.Errorf("bash allowlist missing required entry %q", k)
		}
	}
	for k := range allow {
		if !wantSet[k] {
			t.Errorf("bash allowlist has unexpected extra entry %q", k)
		}
	}
}

// TestEmbeddedAsset_ReporterAgentContract covers the embedded
// vibe-check-reporter.md agent asset: required frontmatter fields, provenance
// marker, required sections, and bash allowlist.
func TestEmbeddedAsset_ReporterAgentContract(t *testing.T) {
	t.Parallel()
	data, err := fs.ReadFile(agentAssetsFS, reporterAgentAssetPath)
	if err != nil {
		t.Fatalf("read embedded reporter agent: %v", err)
	}
	content := string(data)

	// Frontmatter fields.
	for _, needle := range []string{
		"mode: subagent",
		"temperature: 0.3",
		"edit: deny",
		"webfetch: deny",
		`"*": "deny"`,
	} {
		if !strings.Contains(content, needle) {
			t.Errorf("frontmatter missing %q", needle)
		}
	}

	if frontmatterDescription(content) == "" {
		t.Errorf("frontmatter description is empty or missing")
	}

	// Provenance marker.
	if !strings.Contains(content, "<!-- scaffolded by vibe-check") {
		t.Errorf("missing AP-006 provenance marker prefix")
	}

	// Required sections.
	for _, section := range []string{
		"## Source Documents",
		"## Mode Parsing",
		"## Summary Mode",
		"## Detailed Mode",
		"## Trending Mode",
		"## Natural Language Interpretation",
		"## Graceful Degradation",
		"## Security / Operating Constraints",
	} {
		if !strings.Contains(content, "\n"+section) {
			t.Errorf("missing required section heading %q", section)
		}
	}

	// Bash allowlist: exactly 2 entries + catch-all deny.
	allow, denyCatchAll := bashPermissions(content)
	if !denyCatchAll {
		t.Errorf(`bash catch-all "*" must map to "deny"`)
	}
	wantAllow := []string{
		"vibe-check analyze *",
		"git rev-parse *",
	}
	if len(allow) != len(wantAllow) {
		t.Errorf("bash allow count: got %d %v, want %d", len(allow), allow, len(wantAllow))
	}
	for _, k := range wantAllow {
		if !allow[k] {
			t.Errorf("bash allowlist missing required entry %q", k)
		}
	}
	wantSet := make(map[string]bool, len(wantAllow))
	for _, k := range wantAllow {
		wantSet[k] = true
	}
	for k := range allow {
		if !wantSet[k] {
			t.Errorf("bash allowlist has unexpected extra entry %q", k)
		}
	}
}

// TestEmbeddedAsset_CommandContract covers the embedded vibe-check.md command
// asset: required frontmatter fields, provenance marker, mode documentation,
// and $ARGUMENTS passthrough instruction.
func TestEmbeddedAsset_CommandContract(t *testing.T) {
	t.Parallel()
	data, err := fs.ReadFile(commandAssetsFS, commandAssetPath)
	if err != nil {
		t.Fatalf("read embedded command asset: %v", err)
	}
	content := string(data)

	// Frontmatter: description and agent delegation.
	if frontmatterDescription(content) == "" {
		t.Errorf("frontmatter description is empty or missing")
	}
	if !strings.Contains(content, "agent: vibe-check-reporter") {
		t.Errorf("frontmatter missing agent delegation to vibe-check-reporter")
	}

	// Provenance marker.
	if !strings.Contains(content, "<!-- scaffolded by vibe-check") {
		t.Errorf("missing AP-006 provenance marker prefix")
	}

	// Mode documentation: all three modes must be mentioned.
	for _, mode := range []string{"summary", "detailed", "trending"} {
		if !strings.Contains(content, mode) {
			t.Errorf("command body missing mode documentation for %q", mode)
		}
	}

	// $ARGUMENTS passthrough instruction.
	if !strings.Contains(content, "$ARGUMENTS") {
		t.Errorf("command body missing $ARGUMENTS passthrough instruction")
	}
}

// TestAgentAssets_ReturnsReadableFS covers the exported AgentAssets accessor.
func TestAgentAssets_ReturnsReadableFS(t *testing.T) {
	t.Parallel()
	data, err := fs.ReadFile(AgentAssets(), embeddedAgentAssetPath)
	if err != nil {
		t.Fatalf("read via AgentAssets(): %v", err)
	}
	if len(data) == 0 {
		t.Errorf("AgentAssets() returned an empty asset")
	}
}

// TestCommandAssets_ReturnsReadableFS covers the exported CommandAssets
// accessor.
func TestCommandAssets_ReturnsReadableFS(t *testing.T) {
	t.Parallel()
	data, err := fs.ReadFile(CommandAssets(), commandAssetPath)
	if err != nil {
		t.Fatalf("read via CommandAssets(): %v", err)
	}
	if len(data) == 0 {
		t.Errorf("CommandAssets() returned an empty asset")
	}
}

// frontmatterDescription returns the unquoted value of the description field,
// or "" if it is absent. Stdlib scanning only; no YAML dependency.
func frontmatterDescription(content string) string {
	for _, ln := range strings.Split(content, "\n") {
		if strings.HasPrefix(ln, "description:") {
			v := strings.TrimSpace(strings.TrimPrefix(ln, "description:"))
			return strings.Trim(v, `"`)
		}
	}
	return ""
}

// bashPermissions extracts the permission.bash block from the frontmatter,
// returning the set of keys mapped to "allow" and whether the "*" catch-all is
// "deny". Stdlib scanning only; no YAML dependency.
func bashPermissions(content string) (allow map[string]bool, denyCatchAll bool) {
	allow = map[string]bool{}
	inBash := false
	for _, ln := range strings.Split(content, "\n") {
		if !inBash {
			if strings.TrimSpace(ln) == "bash:" {
				inBash = true
			}
			continue
		}
		// Entries are indented four spaces; any dedent ends the block.
		if !strings.HasPrefix(ln, "    ") {
			break
		}
		entry := strings.TrimSpace(ln)
		idx := strings.Index(entry, ":")
		if idx < 0 {
			continue
		}
		key := strings.Trim(strings.TrimSpace(entry[:idx]), `"`)
		val := strings.Trim(strings.TrimSpace(entry[idx+1:]), `"`)
		switch {
		case key == "*":
			denyCatchAll = val == "deny"
		case val == "allow":
			allow[key] = true
		}
	}
	return allow, denyCatchAll
}
