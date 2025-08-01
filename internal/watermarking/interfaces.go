package watermarking

import (
	"encoding/json"
	"io"
)

// Watermarker defines the standard interface for embedding and extracting watermarks.
type Watermarker interface {
	Name() string
	Description() string
	Embed(reader io.Reader, data []byte) (io.Reader, error)
	Extract(reader io.Reader) ([]byte, error)
}

// ConfigParser defines the interface for parsing algorithm-specific parameters
// from raw JSON.
type ConfigParser interface {
	// Parse takes the raw JSON of the "parameters" object and returns a configured Watermarker.
	Parse(params json.RawMessage) (Watermarker, error)
}
