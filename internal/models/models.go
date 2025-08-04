package models

// -- Algorithm listing --
type Algorithm struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// --- Response Structs ---
type HashResponse struct {
	FileUUID string            `json:"file_uuid"`
	Filename string            `json:"filename"`
	Hashes   map[string]string `json:"hashes"`
}

type WaterMarkRespone struct {
	Algorithm string            `json:"algorithm"`
	Watermark map[string]string `json:"watermark"`
}

type FileResponse struct {
	FileUUID   string             `json:"file_uuid"`
	Filename   string             `json:"filename"`
	Hashes     map[string]string  `json:"hashes"`
	Watermarks []WaterMarkHistory `json:"watermarks"`
}

type WaterMarkHistory struct {
	Algorithm string `json:"algorithm"`
	MD5       string `json:"md5"`
}

// HashAlgorithmConfig represents a single hash algorithm and its parameters.
type HashAlgorithmConfig struct {
	Algorithm  string                 `json:"algorithm"`
	Parameters map[string]interface{} `json:"parameters"`
}

// HashConfig represents the config for hash algorithms.
type HashConfig struct {
	HashAlgorithms []HashAlgorithmConfig `json:"hashAlgorithms"`
}

// Struct to define which algorithms to use for the hash value query
type HashValueQueryRequest struct {
	Hashes map[string]string `json:"hashes"`
}

type HashMediaQueryRequest struct {
	Algorithms []string `json:"algorithms"`
}

type EntrySimilarity struct {
	Algorithm  string  `json:"algorithm"`
	UUID       string  `json:"file_uuid"`
	Similarity float64 `json:"similarity"`
	HashId     string  `json:"hash"`
}
