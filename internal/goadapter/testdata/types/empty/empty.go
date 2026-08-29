// Package empty contains no type declarations.
//
// Expected metrics:
//
//	ExportedTypes  = 0
//	AbstractTypes  = 0
package empty

// Version is a package-level constant. Constants are not type declarations
// and do not contribute to ExportedTypes or AbstractTypes counts.
const Version = "1.0.0"
