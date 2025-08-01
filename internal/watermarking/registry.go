package watermarking

import (
	"fmt"
	"mediaguard/internal/models"
)

var registry = make(map[string]Watermarker)

func Register(name string, w Watermarker) {
	registry[name] = w
}

// RegisterWatermarker adds a new watermarker to the registry.
func RegisterWatermarker(wm Watermarker) {
	name := wm.Name()
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("watermarker with name '%s' is already registered", name))
	}
	registry[name] = wm
}

// GetWatermarker retrieves a watermarker from the registry.
func GetWatermarker(name string) (Watermarker, error) {
	wm, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("watermarker '%s' not found", name)
	}
	return wm, nil
}

// ListSupportedAlgorithms returns a slice of all registered hasher names.
func ListSupportedAlgorithms() []models.Algorithm {
	algorithms := make([]models.Algorithm, 0, len(registry))
	for name, watermarker := range registry {

		entry := models.Algorithm{
			Name:        name,
			Description: watermarker.Description(),
		}
		algorithms = append(algorithms, entry)
	}
	return algorithms
}
