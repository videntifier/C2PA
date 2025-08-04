package watermarking

import (
	"io"
)

// Watermarker defines the standard interface for embedding and extracting watermarks.
type Watermarker interface {
	Name() string
	Description() string
	Embed(reader io.Reader, data []byte) (io.Reader, error)
	Extract(reader io.Reader) ([]byte, error)
}
