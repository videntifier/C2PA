package hashing

import (
	"encoding/json"
	"io"
	models "mediaguard/internal/models"
)

// Hasher defines the standard interface for all hashing algorithms.
type Hasher interface {
	Name() string
	Description() string
	ExtractHash(reader io.Reader) (string, error)
	CheckHash(reader io.Reader) ([]models.EntrySimilarity, error)
}

// ConfigParser defines the interface for parsing algorithm-specific parameters
// from raw JSON.
type ConfigParser interface {
	// Parse takes the raw JSON of the "parameters" object and returns a
	// configured Hasher.
	Parse(params json.RawMessage) (Hasher, error)
}
