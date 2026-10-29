package scaffold

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// This file holds the contract tests for the UF policy registration pass:
// the pure splices in splice.go (byte-exact fixture pairs committed under
// testdata/uf/) and the guarded orchestration in register.go exercised
// end-to-end through Run. The asset-writer tests live in scaffold_test.go.

// loadFixture reads a committed UF policy fixture from testdata/uf by name
// and returns its exact content, including any trailing terminator (or its
// absence).
func loadFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "uf", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(data)
}

func TestSpliceCapabilities_InsertsCanonicalEntryInSortedPosition(t *testing.T) {
	t.Parallel()
	reg := registrations[0]

	result := spliceCapabilities([]byte(loadFixture(t, "capabilities-missing.yaml")), reg)

	if !result.inserted {
		t.Errorf("inserted: got false, want true")
	}
	if result.present {
		t.Errorf("present: got true, want false")
	}
	if len(result.notices) != 0 {
		t.Errorf("notices: got %v, want none", result.notices)
	}
	if string(result.content) != loadFixture(t, "capabilities-canonical.yaml") {
		t.Errorf("content: got:\n%s\nwant:\n%s", result.content, loadFixture(t, "capabilities-canonical.yaml"))
	}
}

func TestSpliceCapabilities_SkipsWhenEntryPresent(t *testing.T) {
	t.Parallel()
	reg := registrations[0]
	fixtures := []struct {
		name    string
		content string
	}{
		{"canonical entry", loadFixture(t, "capabilities-canonical.yaml")},
		{"non-canonical scopes", loadFixture(t, "capabilities-non-canonical-scopes.yaml")},
		{"trailing comment", loadFixture(t, "capabilities-entry-trailing-comment.yaml")},
		{"quoted name", loadFixture(t, "capabilities-quoted-name.yaml")},
	}
	for _, tc := range fixtures {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := spliceCapabilities([]byte(tc.content), reg)

			if result.inserted {
				t.Errorf("inserted: got true, want false")
			}
			if !result.present {
				t.Errorf("present: got false, want true")
			}
			if len(result.notices) != 0 {
				t.Errorf("notices: got %v, want none", result.notices)
			}
			if string(result.content) != tc.content {
				t.Errorf("content changed: got:\n%s", result.content)
			}
		})
	}
}

func TestSpliceCapabilities_AdvisoryNoticeForNonReviewPresence(t *testing.T) {
	t.Parallel()
	reg := registrations[0]
	cases := []struct {
		name       string
		content    string
		wantNotice string
	}{
		{
			name:       "non-review capability",
			content:    loadFixture(t, "capabilities-content-capability.yaml"),
			wantNotice: ".uf/reviewer-capabilities.yaml: divisor-entropy present with capability: content; review dispatch may fail closed",
		},
		{
			name:       "empty flow scopes",
			content:    loadFixture(t, "capabilities-empty-flow-scopes.yaml"),
			wantNotice: ".uf/reviewer-capabilities.yaml: divisor-entropy present with empty scopes; review dispatch may fail closed",
		},
		{
			name:       "empty block scopes",
			content:    loadFixture(t, "capabilities-empty-block-scopes.yaml"),
			wantNotice: ".uf/reviewer-capabilities.yaml: divisor-entropy present with empty scopes; review dispatch may fail closed",
		},
		{
			name:       "missing scopes key",
			content:    loadFixture(t, "capabilities-missing-scopes.yaml"),
			wantNotice: ".uf/reviewer-capabilities.yaml: divisor-entropy present with empty scopes; review dispatch may fail closed",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := spliceCapabilities([]byte(tc.content), reg)

			if result.inserted {
				t.Errorf("inserted: got true, want false")
			}
			if !result.present {
				t.Errorf("present: got false, want true")
			}
			if !slices.Equal(result.notices, []string{tc.wantNotice}) {
				t.Errorf("notices: got %v, want [%s]", result.notices, tc.wantNotice)
			}
			if string(result.content) != tc.content {
				t.Errorf("content changed: got:\n%s", result.content)
			}
		})
	}
}

func TestSpliceCapabilities_NoticesWhenReviewersSectionAbsent(t *testing.T) {
	t.Parallel()
	reg := registrations[0]

	result := spliceCapabilities([]byte(loadFixture(t, "capabilities-no-reviewers.yaml")), reg)

	if result.inserted {
		t.Errorf("inserted: got true, want false")
	}
	if result.present {
		t.Errorf("present: got true, want false")
	}
	want := []string{".uf/reviewer-capabilities.yaml: no top-level reviewers: section; entry not inserted"}
	if !slices.Equal(result.notices, want) {
		t.Errorf("notices: got %v, want %v", result.notices, want)
	}
	if string(result.content) != loadFixture(t, "capabilities-no-reviewers.yaml") {
		t.Errorf("content changed: got:\n%s", result.content)
	}
}

func TestSpliceMatrix_InsertsItemInAllModeLists(t *testing.T) {
	t.Parallel()
	reg := registrations[0]

	result := spliceMatrix([]byte(loadFixture(t, "matrix-missing.yaml")), reg)

	if !result.inserted {
		t.Errorf("inserted: got false, want true")
	}
	if result.present {
		t.Errorf("present: got true, want false")
	}
	if len(result.notices) != 0 {
		t.Errorf("notices: got %v, want none", result.notices)
	}
	if string(result.content) != loadFixture(t, "matrix-canonical.yaml") {
		t.Errorf("content: got:\n%s\nwant:\n%s", result.content, loadFixture(t, "matrix-canonical.yaml"))
	}
}

func TestSpliceMatrix_SkipsWhenAllListsPresent(t *testing.T) {
	t.Parallel()
	reg := registrations[0]
	fixtures := []struct {
		name    string
		content string
	}{
		{"canonical items", loadFixture(t, "matrix-canonical.yaml")},
		{"commented and quoted items", loadFixture(t, "matrix-commented-and-quoted-items.yaml")},
	}
	for _, tc := range fixtures {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := spliceMatrix([]byte(tc.content), reg)

			if result.inserted {
				t.Errorf("inserted: got true, want false")
			}
			if !result.present {
				t.Errorf("present: got false, want true")
			}
			if len(result.notices) != 0 {
				t.Errorf("notices: got %v, want none", result.notices)
			}
			if string(result.content) != tc.content {
				t.Errorf("content changed: got:\n%s", result.content)
			}
		})
	}
}

func TestSpliceMatrix_PerListGranularity(t *testing.T) {
	t.Parallel()
	reg := registrations[0]

	result := spliceMatrix([]byte(loadFixture(t, "matrix-partially-registered.yaml")), reg)

	if !result.inserted {
		t.Errorf("inserted: got false, want true (the lists missing the item are spliced)")
	}
	if result.present {
		t.Errorf("present: got true, want false (three lists still gained the item)")
	}
	if len(result.notices) != 0 {
		t.Errorf("notices: got %v, want none", result.notices)
	}
	if string(result.content) != loadFixture(t, "matrix-canonical.yaml") {
		t.Errorf("content: got:\n%s\nwant:\n%s", result.content, loadFixture(t, "matrix-canonical.yaml"))
	}
}

func TestSpliceMatrix_NoticesMissingModeList(t *testing.T) {
	t.Parallel()
	reg := registrations[0]

	result := spliceMatrix([]byte(loadFixture(t, "matrix-no-triage.yaml")), reg)

	if !result.inserted {
		t.Errorf("inserted: got false, want true (remaining lists are spliced)")
	}
	if result.present {
		t.Errorf("present: got true, want false")
	}
	wantNotices := []string{`.uf/review-matrix.yaml: advisor mode list "triage" absent; not created`}
	if !slices.Equal(result.notices, wantNotices) {
		t.Errorf("notices: got %v, want %v", result.notices, wantNotices)
	}
	inserted := 0
	for _, line := range bytes.Split(result.content, []byte("\n")) {
		if string(line) == "    - divisor-entropy" {
			inserted++
		}
	}
	if inserted != 3 {
		t.Errorf("inserted items: got %d, want 3", inserted)
	}
	if bytes.Contains(result.content, []byte("  triage:")) {
		t.Errorf("triage list was created; missing lists must not be created")
	}
}

func TestSpliceMatrix_NoticesWhenNoModeLists(t *testing.T) {
	t.Parallel()
	reg := registrations[0]

	result := spliceMatrix([]byte(loadFixture(t, "matrix-no-mode-lists.yaml")), reg)

	if result.inserted {
		t.Errorf("inserted: got true, want false")
	}
	if result.present {
		t.Errorf("present: got true, want false")
	}
	want := []string{
		`.uf/review-matrix.yaml: advisor mode list "code" absent; not created`,
		`.uf/review-matrix.yaml: advisor mode list "spec" absent; not created`,
		`.uf/review-matrix.yaml: advisor mode list "triage" absent; not created`,
		`.uf/review-matrix.yaml: advisor mode list "feedback" absent; not created`,
	}
	if !slices.Equal(result.notices, want) {
		t.Errorf("notices: got %v, want %v", result.notices, want)
	}
	if string(result.content) != loadFixture(t, "matrix-no-mode-lists.yaml") {
		t.Errorf("content changed: got:\n%s", result.content)
	}
}

func TestRun_RegistersReviewAgentsInPolicyFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir .uf: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ufDir, "reviewer-capabilities.yaml"), []byte(loadFixture(t, "capabilities-missing.yaml")), 0o644); err != nil {
		t.Fatalf("write capabilities fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ufDir, "review-matrix.yaml"), []byte(loadFixture(t, "matrix-missing.yaml")), 0o644); err != nil {
		t.Fatalf("write matrix fixture: %v", err)
	}

	res, err := Run(Options{TargetDir: dir})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !slices.Contains(res.Written, "reviewers/divisor-entropy") {
		t.Errorf("Written missing %q; got %v", "reviewers/divisor-entropy", res.Written)
	}
	if !slices.Contains(res.Written, "advisors/divisor-entropy") {
		t.Errorf("Written missing %q; got %v", "advisors/divisor-entropy", res.Written)
	}
	if slices.Contains(res.Skipped, "reviewers/divisor-entropy") || slices.Contains(res.Skipped, "advisors/divisor-entropy") {
		t.Errorf("Skipped unexpectedly contains registration names; got %v", res.Skipped)
	}
	if len(res.Notices) != 0 {
		t.Errorf("Notices: got %v, want empty", res.Notices)
	}

	capGot, err := os.ReadFile(filepath.Join(ufDir, "reviewer-capabilities.yaml"))
	if err != nil {
		t.Fatalf("read capabilities after Run: %v", err)
	}
	if !bytes.Equal(capGot, []byte(loadFixture(t, "capabilities-canonical.yaml"))) {
		t.Errorf("capabilities after Run: got:\n%s\nwant:\n%s", capGot, loadFixture(t, "capabilities-canonical.yaml"))
	}
	matrixGot, err := os.ReadFile(filepath.Join(ufDir, "review-matrix.yaml"))
	if err != nil {
		t.Fatalf("read matrix after Run: %v", err)
	}
	if !bytes.Equal(matrixGot, []byte(loadFixture(t, "matrix-canonical.yaml"))) {
		t.Errorf("matrix after Run: got:\n%s\nwant:\n%s", matrixGot, loadFixture(t, "matrix-canonical.yaml"))
	}
}

func TestRun_RegistrationSkipsWhenEntriesPresent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir .uf: %v", err)
	}
	capPath := filepath.Join(ufDir, "reviewer-capabilities.yaml")
	matrixPath := filepath.Join(ufDir, "review-matrix.yaml")
	if err := os.WriteFile(capPath, []byte(loadFixture(t, "capabilities-canonical.yaml")), 0o644); err != nil {
		t.Fatalf("write capabilities fixture: %v", err)
	}
	if err := os.WriteFile(matrixPath, []byte(loadFixture(t, "matrix-canonical.yaml")), 0o644); err != nil {
		t.Fatalf("write matrix fixture: %v", err)
	}

	res, err := Run(Options{TargetDir: dir})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !slices.Contains(res.Skipped, "reviewers/divisor-entropy") {
		t.Errorf("Skipped missing %q; got %v", "reviewers/divisor-entropy", res.Skipped)
	}
	if !slices.Contains(res.Skipped, "advisors/divisor-entropy") {
		t.Errorf("Skipped missing %q; got %v", "advisors/divisor-entropy", res.Skipped)
	}
	if slices.Contains(res.Written, "reviewers/divisor-entropy") || slices.Contains(res.Written, "advisors/divisor-entropy") {
		t.Errorf("Written unexpectedly contains registration names; got %v", res.Written)
	}
	if len(res.Notices) != 0 {
		t.Errorf("Notices: got %v, want empty", res.Notices)
	}

	capGot, err := os.ReadFile(capPath)
	if err != nil {
		t.Fatalf("read capabilities after Run: %v", err)
	}
	if !bytes.Equal(capGot, []byte(loadFixture(t, "capabilities-canonical.yaml"))) {
		t.Errorf("capabilities changed: got:\n%s", capGot)
	}
	matrixGot, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read matrix after Run: %v", err)
	}
	if !bytes.Equal(matrixGot, []byte(loadFixture(t, "matrix-canonical.yaml"))) {
		t.Errorf("matrix changed: got:\n%s", matrixGot)
	}
}

func TestRun_ForceLeavesPolicyEntriesUntouched(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir .uf: %v", err)
	}
	capPath := filepath.Join(ufDir, "reviewer-capabilities.yaml")
	matrixPath := filepath.Join(ufDir, "review-matrix.yaml")
	if err := os.WriteFile(capPath, []byte(loadFixture(t, "capabilities-canonical.yaml")), 0o644); err != nil {
		t.Fatalf("write capabilities fixture: %v", err)
	}
	if err := os.WriteFile(matrixPath, []byte(loadFixture(t, "matrix-canonical.yaml")), 0o644); err != nil {
		t.Fatalf("write matrix fixture: %v", err)
	}

	res, err := Run(Options{TargetDir: dir, Force: true})
	if err != nil {
		t.Fatalf("force Run returned error: %v", err)
	}
	if !slices.Contains(res.Skipped, "reviewers/divisor-entropy") {
		t.Errorf("Skipped missing %q; got %v", "reviewers/divisor-entropy", res.Skipped)
	}
	if !slices.Contains(res.Skipped, "advisors/divisor-entropy") {
		t.Errorf("Skipped missing %q; got %v", "advisors/divisor-entropy", res.Skipped)
	}
	if slices.Contains(res.Forced, "reviewers/divisor-entropy") || slices.Contains(res.Forced, "advisors/divisor-entropy") {
		t.Errorf("Forced unexpectedly contains registration names; got %v", res.Forced)
	}

	capGot, err := os.ReadFile(capPath)
	if err != nil {
		t.Fatalf("read capabilities after force Run: %v", err)
	}
	if !bytes.Equal(capGot, []byte(loadFixture(t, "capabilities-canonical.yaml"))) {
		t.Errorf("capabilities changed under force: got:\n%s", capGot)
	}
	matrixGot, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read matrix after force Run: %v", err)
	}
	if !bytes.Equal(matrixGot, []byte(loadFixture(t, "matrix-canonical.yaml"))) {
		t.Errorf("matrix changed under force: got:\n%s", matrixGot)
	}
}

func TestRun_RegistrationAppendsTerminatorAfterUnterminatedFinalLine(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir .uf: %v", err)
	}
	fixture := loadFixture(t, "matrix-unterminated.yaml")
	if err := os.WriteFile(filepath.Join(ufDir, "review-matrix.yaml"), []byte(fixture), 0o644); err != nil {
		t.Fatalf("write matrix fixture: %v", err)
	}

	res, err := Run(Options{TargetDir: dir})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !slices.Contains(res.Written, "advisors/divisor-entropy") {
		t.Errorf("Written missing %q; got %v", "advisors/divisor-entropy", res.Written)
	}

	want := loadFixture(t, "matrix-unterminated-registered.yaml")
	got, err := os.ReadFile(filepath.Join(ufDir, "review-matrix.yaml"))
	if err != nil {
		t.Fatalf("read matrix after Run: %v", err)
	}
	if string(got) != want {
		t.Errorf("matrix after Run: got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRun_RegistrationNoticesWhenUfDirAbsent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	res, err := Run(Options{TargetDir: dir})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	wantNotices := []string{
		".uf/reviewer-capabilities.yaml: not present; not created",
		".uf/review-matrix.yaml: not present; not created",
	}
	if !slices.Equal(res.Notices, wantNotices) {
		t.Errorf("Notices: got %v, want %v", res.Notices, wantNotices)
	}
	if slices.Contains(res.Written, "reviewers/divisor-entropy") || slices.Contains(res.Written, "advisors/divisor-entropy") {
		t.Errorf("Written unexpectedly contains registration names; got %v", res.Written)
	}
	if slices.Contains(res.Skipped, "reviewers/divisor-entropy") || slices.Contains(res.Skipped, "advisors/divisor-entropy") {
		t.Errorf("Skipped unexpectedly contains registration names; got %v", res.Skipped)
	}
	if _, statErr := os.Stat(filepath.Join(dir, ".uf")); !os.IsNotExist(statErr) {
		t.Errorf("expected .uf to remain absent after Run")
	}
}

func TestRun_RegistrationRegistersPresentFileAndNoticesAbsentOne(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir .uf: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ufDir, "review-matrix.yaml"), []byte(loadFixture(t, "matrix-missing.yaml")), 0o644); err != nil {
		t.Fatalf("write matrix fixture: %v", err)
	}

	res, err := Run(Options{TargetDir: dir})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !slices.Contains(res.Written, "advisors/divisor-entropy") {
		t.Errorf("Written missing %q; got %v", "advisors/divisor-entropy", res.Written)
	}
	if slices.Contains(res.Written, "reviewers/divisor-entropy") || slices.Contains(res.Skipped, "reviewers/divisor-entropy") {
		t.Errorf("capabilities outcome unexpected; Written=%v Skipped=%v", res.Written, res.Skipped)
	}
	wantNotices := []string{".uf/reviewer-capabilities.yaml: not present; not created"}
	if !slices.Equal(res.Notices, wantNotices) {
		t.Errorf("Notices: got %v, want %v", res.Notices, wantNotices)
	}

	got, err := os.ReadFile(filepath.Join(ufDir, "review-matrix.yaml"))
	if err != nil {
		t.Fatalf("read matrix after Run: %v", err)
	}
	if !bytes.Equal(got, []byte(loadFixture(t, "matrix-canonical.yaml"))) {
		t.Errorf("matrix after Run: got:\n%s\nwant:\n%s", got, loadFixture(t, "matrix-canonical.yaml"))
	}
}

func TestRun_RegistrationRefusesSymlinkedPolicyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir .uf: %v", err)
	}
	target := filepath.Join(ufDir, "actual-capabilities.yaml")
	if err := os.WriteFile(target, []byte(loadFixture(t, "capabilities-missing.yaml")), 0o644); err != nil {
		t.Fatalf("write symlink target: %v", err)
	}
	if err := os.Symlink(target, filepath.Join(ufDir, "reviewer-capabilities.yaml")); err != nil {
		t.Fatalf("symlink policy file: %v", err)
	}

	_, err := Run(Options{TargetDir: dir})
	if err == nil {
		t.Fatalf("expected error for symlinked policy file, got nil")
	}
	if !strings.Contains(err.Error(), "refusing to overwrite symlink") {
		t.Errorf("error: got %v, want refusal mentioning symlink", err)
	}
	got, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("read symlink target: %v", readErr)
	}
	if !bytes.Equal(got, []byte(loadFixture(t, "capabilities-missing.yaml"))) {
		t.Errorf("symlink target changed: got:\n%s", got)
	}
}

func TestRun_RegistrationRefusesSymlinkedUfDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	realUf := t.TempDir()
	policyTarget := filepath.Join(realUf, "reviewer-capabilities.yaml")
	if err := os.WriteFile(policyTarget, []byte(loadFixture(t, "capabilities-missing.yaml")), 0o644); err != nil {
		t.Fatalf("write policy target: %v", err)
	}
	if err := os.Symlink(realUf, filepath.Join(dir, ".uf")); err != nil {
		t.Fatalf("symlink .uf dir: %v", err)
	}

	_, err := Run(Options{TargetDir: dir})
	if err == nil {
		t.Fatalf("expected error for symlinked .uf dir, got nil")
	}
	if !strings.Contains(err.Error(), "refusing to follow symlink in policy path") {
		t.Errorf("error: got %v, want refusal mentioning symlinked policy path", err)
	}
	got, readErr := os.ReadFile(policyTarget)
	if readErr != nil {
		t.Fatalf("read policy target: %v", readErr)
	}
	if !bytes.Equal(got, []byte(loadFixture(t, "capabilities-missing.yaml"))) {
		t.Errorf("policy target changed: got:\n%s", got)
	}
}

func TestRun_RegistrationRefusesNonRegularPolicyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "reviewer-capabilities.yaml"), 0o755); err != nil {
		t.Fatalf("mkdir non-regular policy path: %v", err)
	}

	_, err := Run(Options{TargetDir: dir})
	if err == nil {
		t.Fatalf("expected error for non-regular policy file, got nil")
	}
	if !strings.Contains(err.Error(), "refusing to write non-regular policy file") {
		t.Errorf("error: got %v, want refusal mentioning non-regular policy file", err)
	}
}

func TestRun_RegistrationPreservesFileMode(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir .uf: %v", err)
	}
	capPath := filepath.Join(ufDir, "reviewer-capabilities.yaml")
	if err := os.WriteFile(capPath, []byte(loadFixture(t, "capabilities-missing.yaml")), 0o600); err != nil {
		t.Fatalf("write capabilities fixture: %v", err)
	}
	if err := os.Chmod(capPath, 0o600); err != nil {
		t.Fatalf("chmod capabilities fixture: %v", err)
	}

	res, err := Run(Options{TargetDir: dir})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !slices.Contains(res.Written, "reviewers/divisor-entropy") {
		t.Errorf("Written missing %q; got %v", "reviewers/divisor-entropy", res.Written)
	}
	fi, err := os.Stat(capPath)
	if err != nil {
		t.Fatalf("stat capabilities after Run: %v", err)
	}
	if got := fi.Mode().Perm(); got != 0o600 {
		t.Errorf("perm after registration: got %o, want %o", got, 0o600)
	}
	got, err := os.ReadFile(capPath)
	if err != nil {
		t.Fatalf("read capabilities after Run: %v", err)
	}
	if !bytes.Equal(got, []byte(loadFixture(t, "capabilities-canonical.yaml"))) {
		t.Errorf("capabilities after Run: got:\n%s", got)
	}
}

func TestRun_RegistrationWriteFailureReturnsPartialWriteError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir .uf: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ufDir, "reviewer-capabilities.yaml"), []byte(loadFixture(t, "capabilities-missing.yaml")), 0o644); err != nil {
		t.Fatalf("write capabilities fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ufDir, "review-matrix.yaml"), []byte(loadFixture(t, "matrix-missing.yaml")), 0o644); err != nil {
		t.Fatalf("write matrix fixture: %v", err)
	}
	sentinel := errors.New("injected write failure")

	_, err := Run(Options{
		TargetDir: dir,
		WriteFile: func(p string, data []byte, perm fs.FileMode) error {
			if filepath.Base(p) == "reviewer-capabilities.yaml" {
				return sentinel
			}
			return os.WriteFile(p, data, perm)
		},
	})
	if err == nil {
		t.Fatalf("expected error from failing WriteFile, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("error chain: got %v, want wrapped %v", err, sentinel)
	}
	if !strings.Contains(err.Error(), "may be partially written") || !strings.Contains(err.Error(), "restore from version control") {
		t.Errorf("error: got %v, want partial-write and restore guidance", err)
	}
}

func TestSpliceCapabilities_NoticesFlowEmptyReviewersSection(t *testing.T) {
	t.Parallel()
	reg := registrations[0]
	content := loadFixture(t, "capabilities-flow-empty-section.yaml")

	result := spliceCapabilities([]byte(content), reg)

	if result.inserted || result.present {
		t.Errorf("outcome: got inserted=%v present=%v, want false/false", result.inserted, result.present)
	}
	wantNotices := []string{
		`.uf/reviewer-capabilities.yaml: reviewers: section carries inline value "[]"; entry not inserted`,
	}
	if !slices.Equal(result.notices, wantNotices) {
		t.Errorf("notices: got %v, want %v", result.notices, wantNotices)
	}
	if string(result.content) != content {
		t.Errorf("content changed: got:\n%s", result.content)
	}
}

func TestSpliceCapabilities_InsertsAfterLastEntryBeforeTrailingComment(t *testing.T) {
	t.Parallel()
	reg := registrations[0]
	content := loadFixture(t, "capabilities-trailing-comment-eof.yaml")

	result := spliceCapabilities([]byte(content), reg)

	if !result.inserted {
		t.Fatalf("inserted: got false, want true")
	}
	if len(result.notices) != 0 {
		t.Errorf("notices: got %v, want none", result.notices)
	}
	want := `# Hand-curated policy whose reviewers section runs to EOF.
reviewers:
  - agent: divisor-entropy
    capability: review
    scopes:
      - standard
  - agent: divisor-guard
    capability: review
    scopes:
      - standard
# scaffolded by uf vdev
`
	if string(result.content) != want {
		t.Errorf("content: got:\n%s\nwant:\n%s", result.content, want)
	}
}

func TestSpliceCapabilities_FallbackIndentWhenNoReviewEntries(t *testing.T) {
	t.Parallel()
	reg := registrations[0]
	content := loadFixture(t, "capabilities-review-less.yaml")

	result := spliceCapabilities([]byte(content), reg)

	if !result.inserted {
		t.Fatalf("inserted: got false, want true")
	}
	if len(result.notices) != 0 {
		t.Errorf("notices: got %v, want none", result.notices)
	}
	want := `# Hand-curated policy whose reviewers declare no review capability.
reviewers:
  - agent: divisor-entropy
    capability: review
    scopes:
      - standard
  - agent: divisor-curator
    capability: content
    scopes:
      - content
`
	if string(result.content) != want {
		t.Errorf("content: got:\n%s\nwant:\n%s", result.content, want)
	}
}

func TestSpliceMatrix_InsertsInPhysicalOrderLists(t *testing.T) {
	t.Parallel()
	reg := registrations[0]
	content := loadFixture(t, "matrix-out-of-order.yaml")

	result := spliceMatrix([]byte(content), reg)

	if !result.inserted {
		t.Fatalf("inserted: got false, want true")
	}
	if result.present {
		t.Errorf("present: got true, want false")
	}
	if len(result.notices) != 0 {
		t.Errorf("notices: got %v, want none", result.notices)
	}
	want := loadFixture(t, "matrix-out-of-order-registered.yaml")
	if string(result.content) != want {
		t.Errorf("content: got:\n%s\nwant:\n%s", result.content, want)
	}

	// A second pass over the registered file must be idempotent: present,
	// byte-identical, with no duplicate items.
	second := spliceMatrix(result.content, reg)
	if second.inserted || !second.present {
		t.Errorf("second pass: got inserted=%v present=%v, want false/true", second.inserted, second.present)
	}
	if string(second.content) != want {
		t.Errorf("second pass content: got:\n%s", second.content)
	}
}

func TestSpliceMatrix_NoticesInlineModeValue(t *testing.T) {
	t.Parallel()
	reg := registrations[0]
	content := loadFixture(t, "matrix-inline-mode.yaml")

	result := spliceMatrix([]byte(content), reg)

	if !result.inserted {
		t.Fatalf("inserted: got false, want true (spec list still gains the item)")
	}
	wantNotices := []string{
		`.uf/review-matrix.yaml: advisor mode list "code" carries inline value "[]"; not edited`,
	}
	if !slices.Equal(result.notices, wantNotices) {
		t.Errorf("notices: got %v, want %v", result.notices, wantNotices)
	}
	want := `# Hand-curated matrix with a flow-empty mode list.
advisor:
  code: []
  spec:
    - divisor-entropy
    - divisor-guard
  triage:
    - divisor-entropy
    - divisor-guard
  feedback:
    - divisor-entropy
    - divisor-guard
`
	if string(result.content) != want {
		t.Errorf("content: got:\n%s\nwant:\n%s", result.content, want)
	}
}

func TestSpliceMatrix_NoticesItemlessModeList(t *testing.T) {
	t.Parallel()
	reg := registrations[0]
	content := loadFixture(t, "matrix-itemless-mode.yaml")

	result := spliceMatrix([]byte(content), reg)

	if !result.inserted {
		t.Fatalf("inserted: got false, want true (spec list still gains the item)")
	}
	wantNotices := []string{
		`.uf/review-matrix.yaml: advisor mode list "code" has no items; not edited`,
	}
	if !slices.Equal(result.notices, wantNotices) {
		t.Errorf("notices: got %v, want %v", result.notices, wantNotices)
	}
	want := `# Hand-curated matrix with an item-less mode key.
advisor:
  code:
  spec:
    - divisor-entropy
    - divisor-guard
  triage:
    - divisor-entropy
    - divisor-guard
  feedback:
    - divisor-entropy
    - divisor-guard
`
	if string(result.content) != want {
		t.Errorf("content: got:\n%s\nwant:\n%s", result.content, want)
	}
}
