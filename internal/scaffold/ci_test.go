package scaffold

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// The fixed pin pair the CI tests render with. The commit is a clearly
// synthetic but well-formed 40-character hex SHA.
const (
	testReleaseTag    = "v0.3.0"
	testReleaseCommit = "0123456789abcdef0123456789abcdef01234567"
)

// ciFixtureDir holds the committed CI scaffolding goldens.
const ciFixtureDir = "testdata/ci"

// updateGoldens rewrites the committed golden files instead of comparing
// them — the standard Go golden-file regeneration switch, matching the
// cmd/vibe-check convention:
//
//	go test ./internal/scaffold/ -update
var updateGoldens = flag.Bool("update", false, "rewrite golden files under testdata/ci instead of comparing")

// Test_assetPaths_ExcludesCIAssets guards design D11: the ci/ subtree has no
// .opencode/ mirror semantics, so assetPaths (which drives the mirror test)
// must never list a ci/-prefixed path.
func Test_assetPaths_ExcludesCIAssets(t *testing.T) {
	t.Parallel()

	paths, err := assetPaths()
	if err != nil {
		t.Fatalf("assetPaths: %v", err)
	}
	for _, p := range paths {
		if strings.HasPrefix(p, "ci/") {
			t.Errorf("assetPaths must exclude the ci/ subtree, got %q (all: %v)", p, paths)
		}
	}
}

// TestCIAssets_ReturnsReadableFS covers the exported CIAssets accessor,
// mirroring the AgentAssets and CommandAssets accessor contracts.
func TestCIAssets_ReturnsReadableFS(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(CIAssets(), ciTemplatePath)
	if err != nil {
		t.Fatalf("read via CIAssets(): %v", err)
	}
	if len(data) == 0 {
		t.Errorf("CIAssets() returned an empty asset")
	}
}

// TestCITemplate_Placeholders guards the template's closed placeholder set
// (design D10): each pin placeholder appears exactly once and no other
// placeholder remains in the embedded template.
func TestCITemplate_Placeholders(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(ciAssetsFS, ciTemplatePath)
	if err != nil {
		t.Fatalf("read embedded CI template: %v", err)
	}
	content := string(data)

	for placeholder, want := range map[string]int{
		"__VIBE_CHECK_VERSION_SHA__": 1,
		"__VIBE_CHECK_VERSION_TAG__": 1,
	} {
		if got := strings.Count(content, placeholder); got != want {
			t.Errorf("placeholder %q count: got %d, want %d", placeholder, got, want)
		}
	}

	for _, line := range strings.Split(content, "\n") {
		for _, field := range strings.Fields(line) {
			if strings.HasPrefix(field, "__") && strings.HasSuffix(field, "__") {
				switch field {
				case "__VIBE_CHECK_VERSION_SHA__", "__VIBE_CHECK_VERSION_TAG__":
					// the two known placeholders, counted above
				default:
					t.Errorf("unknown placeholder %q in template line %q", field, line)
				}
			}
		}
	}
}

// TestValidateReleasePins_Matrix covers the fail-closed pin validation
// (design D6/D12): dev/empty/non-v/garbage/below-floor tags error, missing/
// short/non-hex commit SHAs error, and the floor tag plus newer tags with a
// valid 40-hex SHA pass.
func TestValidateReleasePins_Matrix(t *testing.T) {
	t.Parallel()

	shortSHA := strings.Repeat("a", 16)
	nonHexSHA := strings.Repeat("z", 40)
	longSHA := strings.Repeat("a", 41)

	cases := []struct {
		name       string
		tag        string
		commit     string
		wantErr    bool
		wantSubstr string
	}{
		{name: "dev tag errors", tag: "dev", commit: testReleaseCommit,
			wantErr: true, wantSubstr: `cannot resolve a vibe-check release tag >= v0.3.0 (got "dev")`},
		{name: "empty tag errors", tag: "", commit: testReleaseCommit,
			wantErr: true, wantSubstr: `got ""`},
		{name: "no v prefix errors", tag: "0.3.0", commit: testReleaseCommit,
			wantErr: true, wantSubstr: `got "0.3.0"`},
		{name: "garbage tag errors", tag: "garbage!!", commit: testReleaseCommit,
			wantErr: true, wantSubstr: `got "garbage!!"`},
		{name: "below floor errors", tag: "v0.2.9", commit: testReleaseCommit,
			wantErr: true, wantSubstr: `>= v0.3.0 (got "v0.2.9")`},
		{name: "missing sha errors", tag: "v0.3.0", commit: "",
			wantErr: true, wantSubstr: `cannot resolve the release commit SHA for tag "v0.3.0" (got "")`},
		{name: "short sha errors", tag: "v0.3.0", commit: shortSHA,
			wantErr: true, wantSubstr: `cannot resolve the release commit SHA`},
		{name: "non-hex sha errors", tag: "v0.4.1", commit: nonHexSHA,
			wantErr: true, wantSubstr: `cannot resolve the release commit SHA`},
		{name: "long sha errors", tag: "v0.3.0", commit: longSHA,
			wantErr: true, wantSubstr: `cannot resolve the release commit SHA`},
		{name: "floor tag passes", tag: "v0.3.0", commit: testReleaseCommit},
		{name: "newer tag passes", tag: "v0.4.1", commit: testReleaseCommit},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateReleasePins(tc.tag, tc.commit)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("validateReleasePins(%q, %q): got nil error, want failure", tc.tag, tc.commit)
				}
				if tc.wantSubstr != "" && !strings.Contains(err.Error(), tc.wantSubstr) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.wantSubstr)
				}
			} else if err != nil {
				t.Fatalf("validateReleasePins(%q, %q): got error %v, want nil", tc.tag, tc.commit, err)
			}
		})
	}
}

// TestValidateReleasePins_RemediationMessages verifies both D6 message
// shapes name their remediation, so the operator learns how to fix the pin.
func TestValidateReleasePins_RemediationMessages(t *testing.T) {
	t.Parallel()

	tagErr := validateReleasePins("dev", testReleaseCommit)
	if tagErr == nil {
		t.Fatal("expected tag error, got nil")
	}
	for _, want := range []string{"install vibe-check from a release tag", "pass --vibe-version"} {
		if !strings.Contains(tagErr.Error(), want) {
			t.Errorf("tag error %q missing remediation %q", tagErr.Error(), want)
		}
	}

	shaErr := validateReleasePins("v0.3.0", "nope")
	if shaErr == nil {
		t.Fatal("expected SHA error, got nil")
	}
	for _, want := range []string{"release binary", "network access"} {
		if !strings.Contains(shaErr.Error(), want) {
			t.Errorf("SHA error %q missing remediation %q", shaErr.Error(), want)
		}
	}
}

// TestIsCommitSHA covers the exported 40-character hex predicate.
func TestIsCommitSHA(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		s    string
		want bool
	}{
		{name: "valid lowercase", s: "0123456789abcdef0123456789abcdef01234567", want: true},
		{name: "valid uppercase", s: "0123456789ABCDEF0123456789ABCDEF01234567", want: true},
		{name: "empty", s: "", want: false},
		{name: "short", s: "abc123", want: false},
		{name: "41 chars", s: strings.Repeat("a", 41), want: false},
		{name: "non-hex", s: strings.Repeat("g", 40), want: false},
		{name: "goreleaser none default", s: "none", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := IsCommitSHA(tc.s); got != tc.want {
				t.Errorf("IsCommitSHA(%q): got %v, want %v", tc.s, got, tc.want)
			}
		})
	}
}

// ciGoldenPath returns the committed golden file for the rendered caller
// workflow.
func ciGoldenPath() string {
	return filepath.Join(ciFixtureDir, "vibe-check-structural-gate.golden.yaml")
}

// deployedCIWorkflowPath returns the on-disk location of the deployed caller
// workflow for a given target root.
func deployedCIWorkflowPath(root string) string {
	return filepath.Join(root, ".github", "workflows", CIWorkflowName)
}

// TestRun_CIRendersGoldenWorkflow covers the CI deployment contract (design
// D7/D10/D11): the caller workflow is written under .github/workflows/ with
// 0o644 permissions inside 0o755 directories, reported under the workflows/
// prefix, byte-identical to the committed golden, structurally valid YAML
// with exactly one reusable-workflow job, both pins naming the test release,
// no go-version-file input, and no uncommented threshold inputs.
func TestRun_CIRendersGoldenWorkflow(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	res, err := Run(Options{
		TargetDir:   dir,
		CI:          true,
		VibeVersion: testReleaseTag,
		VibeCommit:  testReleaseCommit,
	})
	if err != nil {
		t.Fatalf("Run with CI returned error: %v", err)
	}
	if !slices.Contains(res.Written, ciWorkflowAssetName) {
		t.Errorf("Written missing %q; got %v", ciWorkflowAssetName, res.Written)
	}
	if slices.Contains(res.Skipped, ciWorkflowAssetName) || slices.Contains(res.Forced, ciWorkflowAssetName) {
		t.Errorf("first CI run must write, not skip or force; got %v / %v / %v", res.Written, res.Skipped, res.Forced)
	}

	wfPath := deployedCIWorkflowPath(dir)
	fi, err := os.Stat(wfPath)
	if err != nil {
		t.Fatalf("stat deployed caller workflow: %v", err)
	}
	if got := fi.Mode().Perm(); got != 0o644 {
		t.Errorf("caller workflow perm: got %o, want %o", got, 0o644)
	}
	for _, dirPermPath := range []string{
		filepath.Join(dir, ".github"),
		filepath.Join(dir, ".github", "workflows"),
	} {
		di, err := os.Stat(dirPermPath)
		if err != nil {
			t.Fatalf("stat %s: %v", dirPermPath, err)
		}
		if !di.IsDir() {
			t.Errorf("%s is not a directory", dirPermPath)
		}
		if got := di.Mode().Perm(); got != 0o755 {
			t.Errorf("%s perm: got %o, want %o", dirPermPath, got, 0o755)
		}
	}

	got, err := os.ReadFile(wfPath)
	if err != nil {
		t.Fatalf("read deployed caller workflow: %v", err)
	}

	goldenPath := ciGoldenPath()
	if *updateGoldens {
		if err := os.MkdirAll(ciFixtureDir, 0o755); err != nil {
			t.Fatalf("mkdir golden dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
			t.Fatalf("rewrite golden: %v", err)
		}
		return
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("rendered caller workflow does not match golden %s\ngot:\n%s\nwant:\n%s", goldenPath, got, want)
	}

	// Structural assertions pin the spec scenarios the golden bytes alone
	// make hard to diagnose: one reusable-workflow job, both pins naming the
	// same test release, language-neutral (no go-version-file), advisory
	// defaults (no uncommented threshold input).
	var doc map[string]any
	if err := yaml.Unmarshal(got, &doc); err != nil {
		t.Fatalf("rendered caller workflow is not valid YAML: %v\n%s", err, got)
	}
	jobs, ok := doc["jobs"].(map[string]any)
	if !ok {
		t.Fatalf("jobs is not a mapping: %T", doc["jobs"])
	}
	if len(jobs) != 1 {
		t.Errorf("caller must contain exactly one job, got %d: %v", len(jobs), jobs)
	}
	job, ok := jobs["structural-gate"].(map[string]any)
	if !ok {
		t.Fatalf("structural-gate job missing or not a mapping: %v", jobs)
	}
	wantUses := "zero-dot-force/vibe-check/.github/workflows/reusable_structural_gate.yml@" + testReleaseCommit
	if got := job["uses"]; got != wantUses {
		t.Errorf("uses pin: got %v, want %q", got, wantUses)
	}
	with, ok := job["with"].(map[string]any)
	if !ok {
		t.Fatalf("with is not a mapping: %T", job["with"])
	}
	if got := with["version"]; got != testReleaseTag {
		t.Errorf("version pin: got %v, want %q", got, testReleaseTag)
	}
	if _, exists := with["go-version-file"]; exists {
		t.Errorf("caller must be language-neutral: got go-version-file %v", with["go-version-file"])
	}
	for _, threshold := range []string{"max-instability-delta", "max-distance-delta", "max-lcom-delta"} {
		if _, exists := with[threshold]; exists {
			t.Errorf("caller must emit advisory defaults: %s must stay commented out", threshold)
		}
	}
	perms, ok := doc["permissions"].(map[string]any)
	if !ok || len(perms) != 1 || perms["contents"] != "read" {
		t.Errorf("caller permissions must be exactly contents: read, got %v", doc["permissions"])
	}
}

// TestRun_CISkipIfPresent covers design D7: an existing caller workflow is
// never overwritten — a rerun reports it as skipped and leaves hand-edited
// content unchanged, with and without --force (force keeps its
// .opencode/-only semantics).
func TestRun_CISkipIfPresent(t *testing.T) {
	t.Parallel()

	for _, force := range []bool{false, true} {
		t.Run(fmt.Sprintf("force=%v", force), func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			if _, err := Run(Options{
				TargetDir:   dir,
				CI:          true,
				VibeVersion: testReleaseTag,
				VibeCommit:  testReleaseCommit,
			}); err != nil {
				t.Fatalf("first CI Run: %v", err)
			}

			// Simulate a hand edit: the user tuned the workflow.
			wfPath := deployedCIWorkflowPath(dir)
			edited := []byte("# hand-edited by the user\nname: mine\n")
			if err := os.WriteFile(wfPath, edited, 0o644); err != nil {
				t.Fatalf("hand-edit caller workflow: %v", err)
			}

			// Rerun with a DIFFERENT pin pair, so a same-content rewrite
			// cannot masquerade as a skip.
			otherCommit := strings.Replace(testReleaseCommit, "0", "1", 1)
			res, err := Run(Options{
				TargetDir:   dir,
				CI:          true,
				Force:       force,
				VibeVersion: "v0.4.1",
				VibeCommit:  otherCommit,
			})
			if err != nil {
				t.Fatalf("rerun CI Run (force=%v): %v", force, err)
			}

			if !slices.Contains(res.Skipped, ciWorkflowAssetName) {
				t.Errorf("rerun Skipped missing %q; got %v", ciWorkflowAssetName, res.Skipped)
			}
			if slices.Contains(res.Written, ciWorkflowAssetName) {
				t.Errorf("rerun must not rewrite the caller workflow; Written=%v", res.Written)
			}
			if slices.Contains(res.Forced, ciWorkflowAssetName) {
				t.Errorf("force must not apply to the caller workflow; Forced=%v", res.Forced)
			}

			got, err := os.ReadFile(wfPath)
			if err != nil {
				t.Fatalf("read caller workflow after rerun: %v", err)
			}
			if !bytes.Equal(got, edited) {
				t.Errorf("hand-edited caller workflow was overwritten:\ngot:\n%s\nwant:\n%s", got, edited)
			}
		})
	}
}

// TestRun_CIValidationErrorAtomicity covers the D6 atomicity guarantee: an
// invalid version with CI set returns an error and writes zero files under
// the target root — no .github/, no .opencode/, nothing.
func TestRun_CIValidationErrorAtomicity(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, err := Run(Options{
		TargetDir:   dir,
		CI:          true,
		VibeVersion: "dev",
		VibeCommit:  testReleaseCommit,
	})
	if err == nil {
		t.Fatal("expected error for invalid version with CI, got nil")
	}
	if !strings.Contains(err.Error(), "install vibe-check from a release tag") {
		t.Errorf("error %q missing the D6 remediation", err.Error())
	}

	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatalf("read target root: %v", readErr)
	}
	if len(entries) != 0 {
		t.Errorf("version error must write zero files, got %v", entries)
	}
}

// TestRun_CINoManifestNotice covers design D8: with no language manifest at
// the target root, the no-manifest notice is reported and the caller
// workflow is still emitted.
func TestRun_CINoManifestNotice(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	res, err := Run(Options{
		TargetDir:   dir,
		CI:          true,
		VibeVersion: testReleaseTag,
		VibeCommit:  testReleaseCommit,
	})
	if err != nil {
		t.Fatalf("Run with CI (no manifest): %v", err)
	}
	if !slices.Contains(res.Notices, noManifestNotice) {
		t.Errorf("Notices missing the no-manifest notice; got %v", res.Notices)
	}
	if _, statErr := os.Stat(deployedCIWorkflowPath(dir)); statErr != nil {
		t.Errorf("caller workflow must still be emitted without a manifest: %v", statErr)
	}
}

// TestRun_CIWithGoModOmitsManifestNotice covers design D8: a go.mod at the
// target root suppresses the no-manifest notice.
func TestRun_CIWithGoModOmitsManifestNotice(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/x\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	res, err := Run(Options{
		TargetDir:   dir,
		CI:          true,
		VibeVersion: testReleaseTag,
		VibeCommit:  testReleaseCommit,
	})
	if err != nil {
		t.Fatalf("Run with CI (go.mod present): %v", err)
	}
	if slices.Contains(res.Notices, noManifestNotice) {
		t.Errorf("Notices must omit the no-manifest notice when go.mod exists; got %v", res.Notices)
	}
}

// TestRun_PlainInitWritesNoGithubTree guards design D1 at the scaffold
// layer: without CI, Run creates nothing under .github/.
func TestRun_PlainInitWritesNoGithubTree(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if _, err := Run(Options{TargetDir: dir}); err != nil {
		t.Fatalf("plain Run: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, ".github")); !os.IsNotExist(statErr) {
		t.Errorf("plain Run must not create .github/: stat err = %v", statErr)
	}
}
