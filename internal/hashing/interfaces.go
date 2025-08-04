package hashing

import (
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
