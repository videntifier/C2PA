package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mediaguard/internal/api"
	"mediaguard/internal/config"
	"mediaguard/internal/database"

	//Import registered hashing algorithms here
	_ "mediaguard/internal/hashing/sha256"
	_ "mediaguard/internal/hashing/vt"

	//Import registered watermarking algorithms here
	_ "mediaguard/internal/watermarking/basic"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection pool
	dbPool, err := database.NewPostgresPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Successfully connected to the database.")

	// Set up the router and API handlers
	router := api.NewRouter(dbPool)

	// Create and start the HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start server: %v", err)
		}
	}()

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting gracefully.")
}
