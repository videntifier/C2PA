
# MediaGuard API Service

MediaGuard is an extendable standalone API service in Go for media hashing and watermarking.

## Table of Contents

- [Features](#features)
- [Assumed Workflows](#assumed-workflows)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Running the Project](#running-the-project)

## Features

-   **Perceptual Hashing**: Generate hashes for images and videos. Currently implements Videntifier Visual Fingerprints and SHA256 Cryptographic hash.
-   **Digital Watermarking**: Embed and extract watermarks. Currently implements metadata embedding via FFMPEG.
-   **Extensible**: Easily add new hashing and watermarking algorithms.
-   **Persistent Storage**: Uses PostgreSQL to store hash and watermarking metadata.
-   **Containerized**: Fully containerized with Docker and orchestrated with Docker Compose.

## Assumed Workflows

### Processing
1. **File arrives** → Send the file for hashing → Hashes are registered in a centralized database.
2. **Send the file for watermarking** → Watermark operation is registered in a centralized database.

### Queries
1. **Submit a file for identification by hash.**
2. **Submit a file for identification by watermark.**

### Challenges to address 
1. **File has been modified.**
2. **File metadata has been stripped.**


## Prerequisites

-   Docker
-   Docker Compose
-   Go (1.24+ for local development/testing)

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

| Variable            | Description                                              | Default Value (from .env.example)                      |
| ------------------- | -------------------------------------------------------- | ------------------------------------------------------ |
| `API_PORT`          | The port on which the API service will listen.           | `8080`                                                 |
| `DB_PORT`           | The port on which the PostgreSQL database will listen.   | `5432`                                                 |
| `POSTGRES_USER`     | Username for the PostgreSQL database.                    | `user`                                                 |
| `POSTGRES_PASSWORD` | Password for the PostgreSQL database.                    | `password`                                             |
| `POSTGRES_DB`       | Name of the PostgreSQL database.                         | `mediaguard`                                           |
| `VSE_ADDRESS`       | Address of the Videntifier Visual Search Engine (VSE).   | `http://vse:7771`                                      |
| `VSE_TOKEN`         | Token for authenticating with the VSE.                   | `VIDENTIFIER`                                          |
| ...other variables  | Additional settings can be added below as needed.        |                                                        |

See env.example for an example environment configuration file

## API Documentation

The API endpoints are documented below. You can use tools like `curl`, Postman, or view the full OpenAPI specification in [`openapi.yaml`](./openapi.yaml).


### `POST /api/v1/hashes`

Calculates and stores perceptual hashes for a media file.

- **Request:** `multipart/form-data`
  - `media`: An image or video file.
  - `config`: (optional) JSON string specifying hash algorithms. If omitted all registered algorithms are used.
- **Response:** `200 OK` with a file UUID and perceptual hashes (pHash, dHash, etc).

### `GET /api/v1/files/{uuid}`

Retrieves all stored information for a file.

- **Response:** `200 OK` with file metadata, hash information, and watermark history.

### `POST /api/v1/watermarks`

Embeds a watermark into a media file.

- **Request:** `multipart/form-data`
  - `media`: The media file.
  - `config`: JSON string specifying watermark algorithm.
  - `data`: The watermark data to embed.
- **Response:** `200 OK` with the watermarked file as an attachment.

### `POST /api/v1/query/hashes/by-media`

Queries the database to find media with similar perceptual hashes.

- **Request:** `multipart/form-data`
  - `media`: The media file to query.
  - `config`: (optional) JSON string specifying hash algorithms.
- **Response:** `200 OK` with a list of matching files.

### `POST /api/v1/query/hashes/by-hash`

Queries the database to find media by hash values.

- **Request:** `application/json`
  - `hashes`: Object mapping algorithm names to hash values.
- **Response:** `200 OK` with a list of matching files.

### `POST /api/v1/query/watermarks`

Extracts a watermark from a media file.

- **Request:** `multipart/form-data`
  - `media`: The media file.
  - `config`: JSON string specifying watermark algorithm.
- **Response:** `200 OK` with extracted watermark data.

### `GET /api/v1/hashes/algorithms`

Lists all supported hash algorithms.

- **Response:** `200 OK` with a list of algorithm names.

### `GET /api/v1/watermarks/algorithms`

Lists all supported watermarking algorithms.

- **Response:** `200 OK` with a list of algorithm names.

### `GET /health`

Health check endpoint.

- **Response:** `200 OK` with body `OK`.


## Extending the API: Adding New Hashing or Watermarking Algorithms

### Adding a New Hashing Algorithm

1. **Create the Algorithm Implementation:**
   - Add your algorithm in a new subfolder under `internal/hashing/{new algo here}`.
   - Implement the required interface defined in `internal/hashing/interfaces.go`.
   - Implement an init function for the algorithm. 
      - Here settings can be read from the environment.
      - In the init function register your algorithm with the hash registry.

2. **Register the Algorithm:**
   - Import your algorithm in `cmd/mediaguard-api/main.go` so your new algorithm is automatically registered and available to the API.

### Adding a New Watermarking Algorithm

1. **Create the Algorithm Implementation:**
   - Add your algorithm in a new subfolder under  `internal/watermarking/{new algo here}`.
   - Implement the required interface defined in `internal/watermarking/interfaces.go`.
   - Implement an init function for the algorithm. 
      - Here settings can be read from the environment.
      - In the init function register your algorithm with the hash registry.

2. **Register the Algorithm:**
   - Import your algorithm in `cmd/mediaguard-api/main.go` so your new algorithm is automatically registered and available to the API.

