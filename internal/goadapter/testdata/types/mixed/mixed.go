// Package mixed provides a variety of type declarations for testing
// abstractness and type-counting metrics.
//
// Expected metrics:
//
//	ExportedTypes  = 7  (Reader, Writer, Point, Config, Pair, Alias, IReader)
//	AbstractTypes  = 2  (Reader, Writer — interfaces)
//
// IReader is an alias to an interface but is classified as CONCRETE per the
// go-adapter spec: aliases never contribute to AbstractTypes.
//
// Unexported types (internal) are excluded from both counts.
package mixed

// Reader is an abstract type for reading bytes.
type Reader interface {
	// Read reads up to len(p) bytes into p.
	Read(p []byte) (n int, err error)
}

// Writer is an abstract type for writing bytes.
type Writer interface {
	// Write writes len(p) bytes from p.
	Write(p []byte) (n int, err error)
}

// Point is a concrete type representing a 2D coordinate.
type Point struct {
	// X is the horizontal coordinate.
	X float64
	// Y is the vertical coordinate.
	Y float64
}

// Config is a concrete type holding configuration values.
type Config struct {
	// Name is the configuration name.
	Name string
	// Debug enables debug mode.
	Debug bool
}

// Pair is a concrete type holding two related values.
type Pair struct {
	// First is the first element.
	First string
	// Second is the second element.
	Second string
}

// Alias is a concrete type alias for string.
type Alias = string

// IReader is a type alias to the exported Reader interface. Although it aliases
// an interface, it MUST be classified as concrete: aliases introduce no new
// abstract type and never contribute to AbstractTypes.
type IReader = Reader

// internal is an unexported struct excluded from exported type counts.
type internal struct { //nolint:unused // exists to test unexported exclusion
	value int
}

// Origin returns a Point at the origin.
func Origin() Point {
	return Point{X: 0, Y: 0}
}
