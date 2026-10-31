package ripwireadapter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ccFile is one File node of ripwire's --export=cc.json output: the
// per-file attribute bundle ripwire reports (CodeCharta shape). Paths are
// joined from the node names below the root folder, "/"-separated,
// matching the file paths the --deps listing reports for the same crawl.
type ccFile struct {
	Path        string
	Loc         int
	Symbols     int
	Cx          int
	CognitiveCx int
	FanIn       int
	FanOut      int
	Churn       int
}

// ccNode is the recursive cc.json node shape.
type ccNode struct {
	Name       string       `json:"name"`
	Type       string       `json:"type"`
	Attributes ccAttributes `json:"attributes"`
	Children   []ccNode     `json:"children"`
}

// ccAttributes carries the per-file metrics; missing keys decode as zero
// (ripwire emits the full set for every File node, but a missing key is
// a data gap, not a structural failure — task 4.3/6.2 consumers treat
// zeros as absent signal).
type ccAttributes struct {
	Loc         int `json:"loc"`
	Symbols     int `json:"symbols"`
	Cx          int `json:"cx"`
	CognitiveCx int `json:"cognitive_cx"`
	FanIn       int `json:"fan_in"`
	FanOut      int `json:"fan_out"`
	Churn       int `json:"churn"`
}

// ccDocument is the cc.json top level. The attributeDescriptors block is
// a stable vocabulary description and is intentionally ignored.
type ccDocument struct {
	ProjectName string   `json:"projectName"`
	APIVersion  string   `json:"apiVersion"`
	Nodes       []ccNode `json:"nodes"`
}

// parseCCOutput parses the --export=cc.json per-file metrics tree. Every
// File node under every root node is flattened into a ccFile with its
// root-relative path. Folder nodes contribute path components only; any
// other node type is ignored. A tree with no File nodes is a valid empty
// result; malformed JSON is an error.
func parseCCOutput(data []byte) ([]ccFile, error) {
	var doc ccDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse cc.json output: %w", err)
	}
	if len(doc.Nodes) == 0 {
		return nil, fmt.Errorf("parse cc.json output: no nodes array (wrong top-level shape)")
	}

	var files []ccFile
	for _, node := range doc.Nodes {
		collectCCFiles(node, "", &files)
	}
	return files, nil
}

// collectCCFiles recursively flattens File nodes, joining path
// components. The synthetic "root" folder name is excluded so paths match
// the --deps listing's root-relative file paths.
func collectCCFiles(node ccNode, prefix string, files *[]ccFile) {
	switch node.Type {
	case "Folder":
		childPrefix := prefix
		if node.Name != "root" && node.Name != "" {
			childPrefix = joinCCPath(prefix, node.Name)
		}
		for _, child := range node.Children {
			collectCCFiles(child, childPrefix, files)
		}
	case "File":
		*files = append(*files, ccFile{
			Path:        joinCCPath(prefix, node.Name),
			Loc:         node.Attributes.Loc,
			Symbols:     node.Attributes.Symbols,
			Cx:          node.Attributes.Cx,
			CognitiveCx: node.Attributes.CognitiveCx,
			FanIn:       node.Attributes.FanIn,
			FanOut:      node.Attributes.FanOut,
			Churn:       node.Attributes.Churn,
		})
	}
}

// joinCCPath joins cc.json path components with "/" separators.
func joinCCPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "/" + name
}

// ccFileByPath indexes parsed files for test and mapping lookups.
func ccFileByPath(files []ccFile) map[string]ccFile {
	byPath := make(map[string]ccFile, len(files))
	for _, f := range files {
		byPath[f.Path] = f
	}
	return byPath
}

// countCCFilesWithPrefix reports how many parsed files live under a path
// prefix — a test convenience for asserting pruned subtrees are absent.
func countCCFilesWithPrefix(files []ccFile, prefix string) int {
	n := 0
	for _, f := range files {
		if strings.HasPrefix(f.Path, prefix) {
			n++
		}
	}
	return n
}
