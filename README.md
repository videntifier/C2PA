
# MediaGuard API Service

MediaGuard is a high-performance, standalone API service in Go for media hashing and watermarking.

## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Running the Project](#running-the-project)

## Features

-   **Perceptual Hashing**: Generate hashes for images and videos using placeholder pHash and dHash algorithms.
-   **Digital Watermarking**: Embed and extract watermarks using a placeholder LSB (Least Significant Bit) watermarking algorithm.
-   **Extensible**: Easily add new hashing and watermarking algorithms.
-   **Persistent Storage**: Uses PostgreSQL to store file metadata and results.
-   **Containerized**: Fully containerized with Docker and orchestrated with Docker Compose.

## Prerequisites

-   Docker
-   Docker Compose
-   Go (1.18+ for local development/testing)

## Getting Started

1.  **Clone or download this project.**

2.  **Download the golang packages**

    ```bash
    go mod tidy 
    ```

2.  **Build and run the services using Docker Compose:**

    The `docker-compose up` command will build the Go application, start the PostgreSQL container, and run the database migrations automatically.

    ```bash
    docker-compose up --build
    ```

The API service will be available at `http://localhost:8080`. The database will be accessible on port `5432`.

## Configuration

The application is configured via environment variables, as defined in `docker-compose.yml`.

| Variable        | Description                                       | Default Value (from docker-compose)                               |
| --------------- | ------------------------------------------------- | ----------------------------------------------------------------- |
| `APP_PORT`      | The port on which the API service will listen.    | `8080`                                                            |
| `DATABASE_URL`  | The connection string for the PostgreSQL database. | `postgres://user:password@postgres-db:5432/mediaguard?sslmode=disable` |

## API Documentation

The API endpoints are documented below. You can use tools like `curl`, Postman, or view the full OpenAPI specification in [`openapi.yaml`](./openapi.yaml).

**Note:** This is a skeleton project. The handlers currently have placeholder logic and do not fully implement file processing.

### `POST /api/v1/hashes`

Calculates and stores perceptual hashes for a media file.

- **Request:** `multipart/form-data`
  - `media`: An image or video file.
- **Response:** `200 OK` with a file UUID and perceptual hashes (pHash, dHash, etc).

### `GET /api/v1/files/{uuid}`

Retrieves all stored information for a file.

- **Response:** `200 OK` with file metadata and hash information.

### `POST /api/v1/watermarks`

Embeds a watermark into a media file.

- **Request:** `multipart/form-data`
  - `media`: The media file.
- **Response:** `200 OK` with watermarking result.

### `POST /api/v1/query/hashes`

Queries the database to find media with similar perceptual hashes.

- **Request:** `application/json`
  - `hash`: The perceptual hash to query for.
- **Response:** `200 OK` with a list of matching files.

### `POST /api/v1/query/watermarks`

Extracts a watermark from a media file.

- **Request:** `multipart/form-data`
  - `media`: The media file.
- **Response:** `200 OK` with extracted watermark data.

---

## OpenAPI Specification

The full API specification is available in [`openapi.yaml`](./openapi.yaml). You can use this file with tools like Swagger UI or Postman to explore and test the API interactively.

---

## Extending the API: Adding New Hashing or Watermarking Algorithms

### Adding a New Hashing Algorithm

1. **Create the Algorithm Implementation:**
   - Add your algorithm in a new file under `internal/hashing/` (or a subfolder).
   - Implement the required interface defined in `internal/hashing/interfaces.go`.

2. **Register the Algorithm:**
   - Update `internal/hashing/registry.go` to register your new algorithm so it is available to the API.

3. **Update Handlers (if needed):**
   - Modify the relevant API handler in `internal/api/handlers.go` to support your new algorithm.

### Adding a New Watermarking Algorithm

1. **Create the Algorithm Implementation:**
   - Add your algorithm in a new file under `internal/watermarking/` (or a subfolder).
   - Implement the required interface defined in `internal/watermarking/interfaces.go`.

2. **Register the Algorithm:**
   - Update `internal/watermarking/registry.go` to register your new algorithm so it is available to the API.

3. **Update Handlers (if needed):**
   - Modify the relevant API handler in `internal/api/handlers.go` to support your new algorithm.

Queries the database to find media with similar perceptual hashes.

-   **Success Response**: `200 OK` (placeholder response).

### `POST /api/v1/query/watermarks`

Extracts a watermark from a media file.

-   **Success Response**: `200 OK` (placeholder response).
