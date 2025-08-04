package hashing

import (
	"fmt"
	"mediaguard/internal/models"
)

var registry = make(map[string]Hasher)

func Register(name string, h Hasher) {
	registry[name] = h
}

// GetHasher retrieves a hasher from the registry by name.
func GetHasher(name string) (Hasher, error) {
	hasher, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("hasher '%s' not found in registry", name)
	}
	return hasher, nil
}

// ListSupportedAlgorithms returns a slice of all registered hasher names.
func ListSupportedAlgorithms() []models.Algorithm {
	algorithms := make([]models.Algorithm, 0, len(registry))
	for name, hasher := range registry {

		entry := models.Algorithm{
			Name:        name,
			Description: hasher.Description(),
		}
		algorithms = append(algorithms, entry)
	}
	return algorithms
}
