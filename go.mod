module mediaguard

go 1.21

replace internal/models => ../internal/models

require (
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/jackc/pgx/v4 v4.17.2
	github.com/joho/godotenv v1.4.0 // For local running outside Docker
)

require (
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.13.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.12.0 // indirect
	github.com/jackc/puddle v1.3.0 // indirect
	golang.org/x/crypto v0.6.0 // indirect
	golang.org/x/text v0.7.0 // indirect
)
