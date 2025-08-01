
package database

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"time"
)

// NewPostgresPool creates a new database connection pool.
func NewPostgresPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	// Configure the connection pool
	config.MaxConns = 10
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	// Ping the database to verify the connection
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, err
	}
	
	log.Println("Database pool created successfully.")
	return pool, nil
}
