// Package ifaces provides interface declarations for testing the go.interfaceWidth
// and go.interfaceProximity extension metrics.
//
// Expected extensions:
//
//	go.interfaceWidth:
//	  Reader:    1  (one method: Read)
//	  Processor: 3  (three methods: Process, Validate, Reset)
//	  Embedder:  2  (embeds Reader + adds Close; flattened = Read + Close)
//
//	go.interfaceProximity:
//	  Reader:    "producer"  (implemented by FileReader in this package)
//	  Processor: "consumer"  (no implementation in this package)
//	  Embedder:  "producer"  (implemented by FileReader in this package)
package ifaces

import "io"

// Reader is a single-method interface for reading data.
// Width = 1.
type Reader interface {
	// Read reads data into p.
	Read(p []byte) (int, error)
}

// Processor is a multi-method interface for data processing.
// Width = 3. No implementation exists in this package (consumer proximity).
type Processor interface {
	// Process processes the input data.
	Process(data []byte) ([]byte, error)
	// Validate checks whether the data is valid.
	Validate(data []byte) error
	// Reset resets the processor to its initial state.
	Reset()
}

// Embedder embeds Reader and adds one method.
// Flattened width = 2 (Read from Reader + Close).
type Embedder interface {
	Reader
	// Close releases resources.
	Close() error
}

// FileReader implements both Reader and Embedder, making them "producer"
// interfaces (implemented in the same package where they are declared).
type FileReader struct {
	path   string    //nolint:unused // used by methods
	reader io.Reader //nolint:unused // used by methods
}

// Read implements Reader.Read and Embedder (via Reader embedding).
func (f *FileReader) Read(p []byte) (int, error) {
	if f.reader == nil {
		return 0, io.EOF
	}
	return f.reader.Read(p)
}

// Close implements Embedder.Close.
func (f *FileReader) Close() error {
	f.reader = nil
	return nil
}

// NewFileReader constructs a FileReader with the given path.
func NewFileReader(path string, r io.Reader) *FileReader {
	return &FileReader{path: path, reader: r}
}
