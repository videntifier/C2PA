package api

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mediaguard/internal/hashing"
	"mediaguard/internal/models"
	"mediaguard/internal/watermarking"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Handlers holds dependencies for HTTP handlers.
type Handlers struct {
	DB *pgxpool.Pool
}

// NewHandlers creates a new Handlers struct.
func NewHandlers(db *pgxpool.Pool) *Handlers {
	return &Handlers{DB: db}
}

// --- Helper Functions ---

// respondWithJSON is a helper to send a JSON response.
func (h *Handlers) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
	}
}

// respondWithError is a helper to send a JSON error message.
func (h *Handlers) respondWithError(w http.ResponseWriter, code int, message string) {
	log.Printf("Error: %s", message)
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

// parseCreateHashesRequest parses the multipart form data to extract the file,
// its content, and the hash configuration.
func (h *Handlers) parseCreateHashesRequest(r *http.Request) (*multipart.FileHeader, *models.HashConfig, []byte, error) {
	file, header, err := r.FormFile("media")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid media file: %w", err)
	}
	defer file.Close()

	var config models.HashConfig
	configStr := r.FormValue("config")
	if configStr != "" {
		if err := json.Unmarshal([]byte(configStr), &config); err != nil {
			return nil, nil, nil, fmt.Errorf("invalid config JSON: %w", err)
		}
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	return header, &config, fileBytes, nil
}

// parseCreateHashesRequest parses the multipart form data to extract the file,
// its content, and the hash configuration.
func (h *Handlers) parseQueryHashesRequest(r *http.Request) (*multipart.FileHeader, *models.HashMediaQueryRequest, []byte, error) {
	file, header, err := r.FormFile("media")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid media file: %w", err)
	}
	defer file.Close()

	var config models.HashMediaQueryRequest
	configStr := r.FormValue("config")
	if configStr != "" {
		if err := json.Unmarshal([]byte(configStr), &config); err != nil {
			return nil, nil, nil, fmt.Errorf("invalid config JSON: %w", err)
		}
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	return header, &config, fileBytes, nil
}

// findExistingFile checks if a file with the given MD5 hash already exists.
func (h *Handlers) findExistingFile(ctx context.Context, md5Hash string) (uuid.UUID, bool, error) {
	var existingFileUUID uuid.UUID
	err := h.DB.QueryRow(
		ctx,
		`SELECT uuid FROM files WHERE md5 = $1`,
		md5Hash,
	).Scan(&existingFileUUID)

	if err == pgx.ErrNoRows {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("db error checking for file: %w", err)
	}
	return existingFileUUID, true, nil
}

// getStoredHashes retrieves the hashes for a given file UUID from the database.
func (h *Handlers) getStoredHashes(ctx context.Context, fileUUID uuid.UUID) (map[string]string, error) {
	hashesFromDB := make(map[string]string)
	rows, err := h.DB.Query(
		ctx,
		`SELECT algorithm, hash_value FROM hashes WHERE file_uuid = $1`,
		fileUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("db error querying hashes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var algo, hashValue string
		if err := rows.Scan(&algo, &hashValue); err != nil {
			return nil, fmt.Errorf("db error scanning hash row: %w", err)
		}
		hashesFromDB[algo] = hashValue
	}
	return hashesFromDB, nil
}

func (h *Handlers) getWatermarkHistory(ctx context.Context, fileUUID uuid.UUID) ([]models.WaterMarkHistory, error) {

	watermarks := make([]models.WaterMarkHistory, 0)

	rows, err := h.DB.Query(
		ctx,
		`SELECT algorithm, md5_after FROM watermark_history WHERE file_uuid = $1`,
		fileUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("db error querying hashes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var watermark models.WaterMarkHistory
		if err := rows.Scan(&watermark.Algorithm, &watermark.MD5); err != nil {
			return nil, fmt.Errorf("db error scanning hash row: %w", err)
		}
		watermarks = append(watermarks, watermark)
	}

	return watermarks, nil

}

// areAllHashesPresent checks if all requested hashes are already in the database.
func areAllHashesPresent(requestedAlgos []models.HashAlgorithmConfig, storedHashes map[string]string) bool {
	for _, algo := range requestedAlgos {
		if _, ok := storedHashes[algo.Algorithm]; !ok {
			return false
		}
	}
	return true
}

// insertFileRecord inserts a new file record and returns its UUID.
func (h *Handlers) insertFileRecord(ctx context.Context, filename, mediaType, md5Hash string) (uuid.UUID, error) {
	var fileUUID uuid.UUID
	err := h.DB.QueryRow(
		ctx,
		`INSERT INTO files (filename, media_type, md5) VALUES ($1, $2, $3) RETURNING uuid`,
		filename,
		mediaType,
		md5Hash,
	).Scan(&fileUUID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("db error inserting file: %w", err)
	}
	return fileUUID, nil
}

func (h *Handlers) insertWatermarkRecord(ctx context.Context, file_uuid uuid.UUID, md5, algorithm string) error {

	_, err := h.DB.Exec(
		ctx,
		`INSERT INTO watermark_history(file_uuid, algorithm, md5_after) VALUES ($1, $2, $3)`,
		file_uuid,
		algorithm,
		md5)

	if err != nil {
		return fmt.Errorf("db error inserting file: %w", err)
	}
	return nil

}

// generateAndStoreHashes generates hashes for the given file and stores them.
func (h *Handlers) generateAndStoreHashes(ctx context.Context, fileBytes []byte, fileUUID uuid.UUID, algos []models.HashAlgorithmConfig) (map[string]string, error) {
	hashValues := make(map[string]string)
	for _, algo := range algos {
		log.Printf("Configured hasher: %s with parameters: %+v", algo.Algorithm, algo.Parameters)
		hasher, err := hashing.GetHasher(algo.Algorithm)
		if err != nil {
			return nil, fmt.Errorf("invalid config for hasher %s: %w", algo.Algorithm, err)
		}

		hashValue, err := hasher.ExtractHash(bytes.NewReader(fileBytes))
		if err != nil {
			return nil, fmt.Errorf("failed hashing with %s: %w", algo.Algorithm, err)
		}

		log.Printf("%s : %s", algo.Algorithm, hashValue)
		hashValues[algo.Algorithm] = hashValue

		_, err = h.DB.Exec(
			ctx,
			`INSERT INTO hashes (file_uuid, algorithm, hash_value) VALUES ($1, $2, $3) ON CONFLICT (file_uuid, algorithm) DO NOTHING`,
			fileUUID,
			algo.Algorithm,
			hashValue,
		)
		if err != nil {
			return nil, fmt.Errorf("db error inserting hash for %s: %w", algo.Algorithm, err)
		}
	}
	return hashValues, nil
}

// HandleCreateHashes handles the creation of hashes for a media file.
func (h *Handlers) HandleCreateHashes(w http.ResponseWriter, r *http.Request) {
	header, config, fileBytes, err := h.parseCreateHashesRequest(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	md5Hasher := md5.New()
	md5Hasher.Write(fileBytes)
	md5Hash := hex.EncodeToString(md5Hasher.Sum(nil))

	fileUUID, found, err := h.findExistingFile(r.Context(), md5Hash)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error while searching for file.")
		return
	}

	if found {
		storedHashes, err := h.getStoredHashes(r.Context(), fileUUID)
		if err != nil {
			h.respondWithError(w, http.StatusInternalServerError, "Database error while getting stored hashes.")
			return
		}

		if areAllHashesPresent(config.HashAlgorithms, storedHashes) {
			response := models.HashResponse{
				FileUUID: fileUUID.String(),
				Filename: filepath.Base(header.Filename),
				Hashes:   storedHashes,
			}
			h.respondWithJSON(w, http.StatusOK, response)
			return
		}
	}

	if !found {
		fileUUID, err = h.insertFileRecord(r.Context(), header.Filename, header.Header.Get("Content-Type"), md5Hash)
		if err != nil {
			h.respondWithError(w, http.StatusInternalServerError, "Database error while inserting new file record.")
			return
		}
	}

	_, err = h.generateAndStoreHashes(r.Context(), fileBytes, fileUUID, config.HashAlgorithms)
	if err != nil {
		if strings.Contains(err.Error(), "invalid config") {
			h.respondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "Error during hashing process.")
		}
		return
	}

	// Refetch all hashes to ensure the response is complete
	allHashes, err := h.getStoredHashes(r.Context(), fileUUID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error while retrieving final hashes.")
		return
	}

	response := models.HashResponse{
		FileUUID: fileUUID.String(),
		Filename: filepath.Base(header.Filename),
		Hashes:   allHashes,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// --- Other Handlers ---

// Lists all the available hashes for this file
func (h *Handlers) HandleListMediaHashes(w http.ResponseWriter, r *http.Request) {
	// Get the uuid from the URL
	vars := mux.Vars(r)
	uuidStr, ok := vars["uuid"]
	if !ok {
		h.respondWithError(w, http.StatusBadRequest, "File UUID is required")
		return
	}

	fileUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid file UUID format")
		return
	}

	// Get the file information from the database
	var filename string
	err = h.DB.QueryRow(r.Context(), `SELECT filename FROM files WHERE uuid = $1`, fileUUID).Scan(&filename)
	if err != nil {
		if err == pgx.ErrNoRows {
			h.respondWithError(w, http.StatusNotFound, "File not found")
			return
		}
		h.respondWithError(w, http.StatusInternalServerError, "Database error while retrieving file information")
		return
	}

	// Get the hash information from the database
	hashes, err := h.getStoredHashes(r.Context(), fileUUID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error while retrieving hashes")
		return
	}

	// Get the watermark information from the database
	watermarks, err := h.getWatermarkHistory(r.Context(), fileUUID)

	// Return json with hash list
	response := models.FileResponse{
		FileUUID:   fileUUID.String(),
		Filename:   filepath.Base(filename),
		Hashes:     hashes,
		Watermarks: watermarks,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// HandleEmbedWatermark embeds a watermark into the provided media file.
func (h *Handlers) HandleEmbedWatermark(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form data
	file, header, err := r.FormFile("media")
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid media file: "+err.Error())
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Failed to read file: "+err.Error())
		return
	}

	// Parse algorithm config
	configStr := r.FormValue("config")
	if configStr == "" {
		h.respondWithError(w, http.StatusBadRequest, "Missing watermark algorithm config")
		return
	}
	var config struct {
		Algorithm string `json:"algorithm"`
	}
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid config JSON: "+err.Error())
		return
	}

	// Parse watermark data
	dataStr := r.FormValue("data")
	if dataStr == "" {
		h.respondWithError(w, http.StatusBadRequest, "Missing watermark data")
		return
	}

	watermarkData := []byte(dataStr)

	//Check if we have this file already registered based on the MD5
	md5Hasher := md5.New()
	md5Hasher.Write(fileBytes)
	md5Hash := hex.EncodeToString(md5Hasher.Sum(nil))

	fileUUID, found, err := h.findExistingFile(r.Context(), md5Hash)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error while searching for file.")
		return
	}

	//We have not seen this file before so we register it
	if !found {
		fileUUID, err = h.insertFileRecord(r.Context(), header.Filename, header.Header.Get("Content-Type"), md5Hash)
		if err != nil {
			h.respondWithError(w, http.StatusInternalServerError, "Database error while inserting new file record.")
			return
		}
	}

	// Get the watermarker
	watermarker, err := watermarking.GetWatermarker(config.Algorithm)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Unknown watermark algorithm: "+err.Error())
		return
	}

	// Embed watermark
	resultBytes, err := watermarker.Embed(bytes.NewReader(fileBytes), watermarkData)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to embed watermark: "+err.Error())
		return
	}

	resultBytesData, err := io.ReadAll(resultBytes)

	//Compute the MD5 of the watermarked file
	md5Hasher = md5.New()
	md5Hasher.Write(resultBytesData)

	//Register the watermark action with the database
	err = h.insertWatermarkRecord(r.Context(), fileUUID, hex.EncodeToString(md5Hasher.Sum(nil)), watermarker.Name())

	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to register watermark"+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="watermarked_`+filepath.Base(header.Filename)+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resultBytesData)

	// Prepare multipart/mixed response
	/*var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	w.Header().Set("Content-Type", "multipart/mixed; boundary="+mw.Boundary())

	// Part 1: JSON metadata
	jsonPart, _ := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type": []string{"application/json"},
	})

	metadata := map[string]interface{}{
		"file_uuid": fileUUID.String(),
		"algorithm": watermarker.Name(),
	}

	_ = json.NewEncoder(jsonPart).Encode(metadata)

	// Part 2: Watermarked file
	filePart, _ := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":        []string{"application/octet-stream"},
		"Content-Disposition": []string{`attachment; filename="watermarked_` + filepath.Base(header.Filename) + `"`},
	})
	_, _ = filePart.Write(resultBytesData)*/
}

func (h *Handlers) HandleExtractWatermark(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form data
	file, _, err := r.FormFile("media")
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid media file: "+err.Error())
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Failed to read file: "+err.Error())
		return
	}

	// Parse algorithm config
	configStr := r.FormValue("config")
	if configStr == "" {
		h.respondWithError(w, http.StatusBadRequest, "Missing watermark algorithm config")
		return
	}
	var config struct {
		Algorithm string `json:"algorithm"`
	}
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid config JSON: "+err.Error())
		return
	}

	// Get the watermarker
	watermarker, err := watermarking.GetWatermarker(config.Algorithm)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Unknown watermark algorithm: "+err.Error())
		return
	}

	// Embed watermark
	resultBytes, err := watermarker.Extract(bytes.NewReader(fileBytes))
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to extract watermark: "+err.Error())
		return
	}

	var watermarkMap map[string]string
	if err := json.Unmarshal(resultBytes, &watermarkMap); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to parse extracted watermark: "+err.Error())
		return
	}

	// Return json with hash list
	response := models.WaterMarkRespone{
		Algorithm: watermarker.Name(),
		Watermark: watermarkMap,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *Handlers) HandleQueryHashesByMedia(w http.ResponseWriter, r *http.Request) {

	//Parse the request data
	_, config, fileBytes, err := h.parseQueryHashesRequest(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	//If the config is empty add all available algorithm names to config
	if len(config.Algorithms) == 0 {
		algos := hashing.ListSupportedAlgorithms()

		for _, algo := range algos {
			config.Algorithms = append(config.Algorithms, algo.Name)
		}
	}

	results := make([]models.EntrySimilarity, 0)
	//Iterate over all the algorith types found in the config
	for _, algo := range config.Algorithms {

		//Get the correct hasher
		hasher, err := hashing.GetHasher(algo)
		if err != nil {
			log.Printf("[ERROR] Failed to get hasher for algorithm %s: %v", algo, err)
			h.respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Failed to get hasher for algorithm %s", algo))
			return
		}

		entries, err := hasher.CheckHash(bytes.NewReader(fileBytes))

		if err != nil {
			log.Printf("[ERROR] failed to check hashes for algorithm %s: %v", algo, err)
			h.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check hashes for algorithm %s", algo))
			return
		}

		//Iterate over the entries and get the linked file-id
		for idx := range entries {
			err = h.getFileUUIDByHash(r.Context(), &entries[idx])

			if err != nil {
				// If no rows are found, it's not an error we need to log verbosely.
				// The entry will just have an empty UUID and be filtered out.
				if err != pgx.ErrNoRows {
					log.Printf("[ERROR] Failed to find linked uuid for hash %s: %v", entries[idx].HashId, err)
				}
			}
		}

		//Remove all entries in entries with empty UUID
		filteredEntries := make([]models.EntrySimilarity, 0, len(entries))
		for _, entry := range entries {
			if entry.UUID != "" {
				filteredEntries = append(filteredEntries, entry)
			}
		}

		results = append(results, filteredEntries...)
	}

	h.respondWithJSON(w, http.StatusOK, results)
}

func (h *Handlers) getFileUUIDByHash(ctx context.Context, entry *models.EntrySimilarity) error {
	const FILEQUERY = `SELECT file_uuid FROM hashes WHERE algorithm = $1 AND hash_value = $2`
	var fileUUID uuid.UUID
	err := h.DB.QueryRow(ctx, FILEQUERY, entry.Algorithm, entry.HashId).Scan(&fileUUID)
	if err != nil {
		return err
	}
	entry.UUID = fileUUID.String()
	return nil
}

func (h *Handlers) HandleQueryHashesByHashValue(w http.ResponseWriter, r *http.Request) {

	// Parse the request body into HashValueQueryRequest
	var req models.HashValueQueryRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	results := make([]*models.EntrySimilarity, 0)

	for algo, hashValue := range req.Hashes {
		hasher, err := hashing.GetHasher(algo)
		if err != nil {
			log.Printf("Unknown hasher: %s", algo)
			continue // skip unknown hashers
		}

		entry, err := h.GetEntryByAlgorithmAndHash(r.Context(), hasher.Name(), hashValue)
		if err != nil {
			log.Printf("Error checking hash for %s: %v", algo, err)
			continue // skip errors
		}

		if entry != nil {
			results = append(results, entry)
		}
	}

	h.respondWithJSON(w, http.StatusOK, results)
}

// HandleHashAlgorithmListing returns a list of supported hash algorithms.
func (h *Handlers) HandleHashAlgorithmListing(w http.ResponseWriter, r *http.Request) {
	algorithms := hashing.ListSupportedAlgorithms()
	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{"algorithms": algorithms})
}

// HandleWatermarkAlgorithmListing returns a list of supported watermarking algorithms.
func (h *Handlers) HandleWatermarkAlgorithmListing(w http.ResponseWriter, r *http.Request) {
	watermarks := watermarking.ListSupportedAlgorithms()
	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{"algorithms": watermarks})
}

// GetEntryByAlgorithmAndHash queries the database for an entry matching the algorithm and hash value.
func (h *Handlers) GetEntryByAlgorithmAndHash(ctx context.Context, algorithm string, hashValue string) (*models.EntrySimilarity, error) {

	row := h.DB.QueryRow(ctx, `SELECT file_uuid, algorithm, hash_value FROM hashes WHERE algorithm = $1 AND hash_value = $2 LIMIT 1`, algorithm, hashValue)
	var fileUUID, algo, hash string
	if err := row.Scan(&fileUUID, &algo, &hash); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &models.EntrySimilarity{
		Algorithm:  algo,
		UUID:       fileUUID,
		Similarity: 100,
		HashId:     hash,
	}, nil
}
