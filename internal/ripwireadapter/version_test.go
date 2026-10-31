package ripwireadapter

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

// Version-pin tests — openspec/changes/ripwire-adapter task 2.2.
//
// They exercise the fail-closed pin through the production path
// (Adapter.Analyze → exec.LookPath → --version subprocess → parse →
// compare) using the fake-ripwire harness: wrapper installs select the
// version-output variant, and a PATH without the fake exercises the
// absent-binary guidance. Every failure path must return a nil graph.

const (
	wrongVersionLine = "ripwire 0.6.8 (Release, AppleClang 21.0.0.21000101, emit=std::print, built_from=deadbeef00)"
	garbageLine      = "totally not a version string"
)

func TestAnalyze_PinGate_AbsentBinary(t *testing.T) {
	// A PATH with no ripwire (real or fake) anywhere on it.
	t.Setenv("PATH", t.TempDir())

	graph, err := NewPython().Analyze(t.Context(), ".")
	if graph != nil {
		t.Fatalf("absent binary: got graph %v, want nil", graph)
	}
	if !errors.Is(err, ErrRipwireAbsent) {
		t.Fatalf("absent binary: got error %v, want ErrRipwireAbsent", err)
	}
	for _, want := range []string{"ripwire", ripwireInstallURL, "0.6.9", "bedc0cd"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("absent-binary guidance: error %q does not name %q", err, want)
		}
	}
}

func TestAnalyze_PinGate_ThroughProductionPath(t *testing.T) {
	// Not parallel: the wrapper installs pin process-wide PATH state.
	cases := []struct {
		name      string
		variant   string
		wantError error
		wantText  []string
	}{
		{
			name:      "matching version passes pin",
			variant:   "valid",
			wantError: nil, // the not-implemented stub, not a pin error
		},
		{
			name:      "wrong version fails closed",
			variant:   "wrong-version",
			wantError: ErrRipwireVersionMismatch,
			wantText:  []string{"0.6.9", "bedc0cd", "0.6.8", "deadbeef00"},
		},
		{
			name:      "garbage version fails closed",
			variant:   "garbage",
			wantError: ErrRipwireVersionUnparseable,
			wantText:  []string{garbageLine},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recordings := filepath.Join("testdata", "recorded", "python")
			installFakeRipwireWrapper(t, recordings, tc.variant)

			graph, err := NewPython().Analyze(t.Context(), ".")
			if tc.wantError == nil {
				// The pin passed: analysis proceeds and produces a
				// graph, with no pin sentinel anywhere in the path.
				if err != nil {
					t.Fatalf("valid version: unexpected error %v", err)
				}
				if graph == nil {
					t.Fatal("valid version: got nil graph, want assembled graph")
				}
				if len(graph.Modules) == 0 {
					t.Fatal("valid version: got empty graph")
				}
				return
			}
			if graph != nil {
				t.Fatalf("got graph %v, want nil", graph)
			}
			if !errors.Is(err, tc.wantError) {
				t.Fatalf("got error %v, want %v", err, tc.wantError)
			}
			for _, want := range tc.wantText {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not name %q", err, want)
				}
			}
		})
	}
}

func TestParseRipwireVersion(t *testing.T) {
	t.Parallel()

	recorded := readRecording(t, filepath.Join("testdata", "recorded", "python", "version.txt"))

	cases := []struct {
		name        string
		output      string
		wantVersion string
		wantCommit  string
		wantErr     bool
	}{
		{
			name:        "recorded validated line",
			output:      string(recorded),
			wantVersion: "0.6.9",
			wantCommit:  "bedc0cdcf",
		},
		{
			name:        "wrong version line still parses",
			output:      wrongVersionLine,
			wantVersion: "0.6.8",
			wantCommit:  "deadbeef00",
		},
		{
			name:    "garbage output",
			output:  garbageLine,
			wantErr: true,
		},
		{
			name:    "empty output",
			output:  "",
			wantErr: true,
		},
		{
			name:    "missing built_from commit",
			output:  "ripwire 0.6.9 (Release)",
			wantErr: true,
		},
		{
			name:    "version word only",
			output:  "ripwire",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			info, err := parseRipwireVersion([]byte(tc.output))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parse %q: got %v, want error", tc.output, info)
				}
				if !errors.Is(err, ErrRipwireVersionUnparseable) {
					t.Errorf("parse %q: got error %v, want ErrRipwireVersionUnparseable", tc.output, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse %q: %v", tc.output, err)
			}
			if info.version != tc.wantVersion {
				t.Errorf("version: got %q, want %q", info.version, tc.wantVersion)
			}
			if info.commit != tc.wantCommit {
				t.Errorf("commit: got %q, want %q", info.commit, tc.wantCommit)
			}
		})
	}
}

func TestCommitPrefixMatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		version string
		commit  string
		want    bool
	}{
		{"recorded commit extends pin", "0.6.9", "bedc0cdcf", true},
		{"exact pin commit", "0.6.9", "bedc0cd", true},
		// bedc0cdff also extends the 7-character pinned prefix; the
		// prefix pin accepts it (a collision needs 28 matching bits,
		// and the exact binary is additionally sha256-verified in the
		// testdata README).
		{"another commit extending the pin prefix", "0.6.9", "bedc0cdff", true},
		{"commit diverging inside the pin prefix", "0.6.9", "bedc0cef", false},
		{"unrelated commit", "0.6.9", "deadbeef00", false},
		{"right commit wrong version", "0.6.8", "bedc0cdcf", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			info := ripwireVersionInfo{version: tc.version, commit: tc.commit}
			if got := info.matchesPin(); got != tc.want {
				t.Errorf("matchesPin(%s @ %s): got %v, want %v", tc.version, tc.commit, got, tc.want)
			}
		})
	}
}

func TestParseRipwireVersion_OversizeOutput(t *testing.T) {
	t.Parallel()

	output := make([]byte, maxVersionOutputBytes+1)
	for i := range output {
		output[i] = 'x'
	}
	_, err := parseRipwireVersion(output)
	if !errors.Is(err, ErrRipwireVersionUnparseable) {
		t.Fatalf("oversize output: got error %v, want ErrRipwireVersionUnparseable", err)
	}
}

// TestPinGate_SanitizedEnvironment proves the production pin check carries
// no host credentials to the subprocess and no FAKE_RIPWIRE_* passthrough:
// the fake works purely through the wrapper's baked-in control variables.
func TestPinGate_SanitizedEnvironment(t *testing.T) {
	recordings := filepath.Join("testdata", "recorded", "typescript")
	installFakeRipwireWrapper(t, recordings, "valid")

	// Host credential that must never reach the subprocess.
	t.Setenv("GITHUB_TOKEN", "definitely-a-secret")

	graph, err := NewTypeScript().Analyze(t.Context(), ".")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if graph == nil {
		t.Fatal("got nil graph, want assembled graph")
	}
	// The sanitized graph output must not carry the host credential either.
	encoded, marshalErr := json.Marshal(graph)
	if marshalErr != nil {
		t.Fatalf("marshal graph: %v", marshalErr)
	}
	if strings.Contains(string(encoded), "definitely-a-secret") {
		t.Fatal("graph output leaked a host credential")
	}
}
