package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
)

// corsMiddleware adds CORS headers to each response
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//TODO Here adjust the origins to the appropriate sources
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// NewRouter creates and configures a new application router.
func NewRouter(db *pgxpool.Pool) *mux.Router {
	router := mux.NewRouter()

	// Add CORS middleware to all routes
	router.Use(corsMiddleware)

	// Create a subrouter for API versioning
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// Instantiate handlers
	h := NewHandlers(db)

	// API Endpoints
	apiV1.HandleFunc("/hashes", h.HandleCreateHashes).Methods(http.MethodPost)
	apiV1.HandleFunc("/hashes/algorithms", h.HandleHashAlgorithmListing).Methods(http.MethodGet)

	apiV1.HandleFunc("/query/hashes/by-media", h.HandleQueryHashesByMedia).Methods(http.MethodPost)
	apiV1.HandleFunc("/query/hashes/by-hash", h.HandleQueryHashesByHashValue).Methods(http.MethodPost)
	apiV1.HandleFunc("/query/hashes/by-mpd-playlist", h.HandleQueryHashesByMPDPlaylist).Methods(http.MethodPost)

	apiV1.HandleFunc("/files/{uuid:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}", h.HandleListMediaHashes).Methods(http.MethodGet)

	apiV1.HandleFunc("/watermarks", h.HandleEmbedWatermark).Methods(http.MethodPost)
	apiV1.HandleFunc("/query/watermarks", h.HandleExtractWatermark).Methods(http.MethodPost)
	apiV1.HandleFunc("/watermarks/algorithms", h.HandleWatermarkAlgorithmListing).Methods(http.MethodGet)

	// Add a simple health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	return router
}
