package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// This file implements the registration pass from the
// init-registers-divisor-entropy change (design D1-D10): after deploying
// assets, init splices the registered review agent into the target
// project's Unbound Force policy files. The pure content splices live in
// splice.go and their line-scanning helpers in lines.go; this file holds
// the registration table and the guarded filesystem orchestration.

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
