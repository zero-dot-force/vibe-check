package metrics

// Duplication represents a set of structurally similar code blocks within a
// module. Two or more blocks are considered duplicates when their normalized
// AST subtrees are deeply equal after identifier and literal normalization.
//
// The Similarity field ranges from 0.0 (no similarity) to 1.0 (identical
// structure after normalization). A value of 1.0 indicates the blocks differ
// only in identifier names and/or literal values.
//
// Duplication is language-agnostic: the Blocks field carries file paths and
// line ranges, while the detection logic lives in language-specific adapters.
type Duplication struct {
	// ModulePath is the unique identifier for the module containing the
	// duplicate blocks (e.g., "github.com/foo/bar").
	ModulePath string `json:"modulePath"`

	// Blocks contains the locations of the duplicate code blocks. There
	// must be at least two blocks for a valid duplication.
	Blocks []DuplicateBlock `json:"blocks"`

	// Similarity is the structural similarity score between the blocks.
	// Range: [0.0, 1.0]. A value of 1.0 means the blocks are structurally
	// identical after identifier and literal normalization.
	Similarity float64 `json:"similarity"`
}

// DuplicateBlock describes the location and size of a single duplicate code
// block within a source file. Line numbers are 1-indexed and inclusive.
type DuplicateBlock struct {
	// File is the path to the source file containing this block, relative
	// to the project root.
	File string `json:"file"`

	// StartLine is the 1-indexed line number where the duplicate block
	// begins. Must be >= 1.
	StartLine int `json:"startLine"`

	// EndLine is the 1-indexed line number where the duplicate block ends
	// (inclusive). Must be >= StartLine.
	EndLine int `json:"endLine"`

	// LineCount is the number of significant lines in this block (excluding
	// blank lines and single-token lines such as solitary braces). Must be
	// >= 0. A block with LineCount < the minimum threshold (typically 6) is
	// excluded from duplication detection.
	LineCount int `json:"lineCount"`
}
