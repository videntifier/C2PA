package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config stores all configuration for the application.
type Config struct {
	Port        string
	DatabaseURL string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	// Load .env file if it exists (useful for local development without Docker)
	godotenv.Load()

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080" // Default port
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Provide a default for local dev if needed
		dbURL = "postgres://user:password@localhost:5432/mediaguard?sslmode=disable"
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
	}, nil
}
