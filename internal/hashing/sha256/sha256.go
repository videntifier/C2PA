package sha256

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mediaguard/internal/hashing"
	"mediaguard/internal/models"
)

type SHA256Hash struct{}
type SHA256Parser struct{}
type SHA256Config struct{}

const AlgorithmSHA256 = "sha256"

func init() {
	hashing.Register(AlgorithmSHA256, &SHA256Hash{})
}

func (h *SHA256Hash) Name() string {
	return AlgorithmSHA256
}

func (h *SHA256Hash) Description() string {
	return "SHA-256 is a cryptographic hash function that produces a fixed-size 256-bit (32-byte) hash values"
}

func (h *SHA256Hash) ExtractHash(reader io.Reader) (string, error) {

	hash := sha256.New()

	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (h *SHA256Hash) CheckHash(reader io.Reader) ([]models.EntrySimilarity, error) {

	var entries []models.EntrySimilarity

	hash := sha256.New()

	if _, err := io.Copy(hash, reader); err != nil {
		return entries, err
	}

	var entry models.EntrySimilarity
	entry.Algorithm = AlgorithmSHA256
	entry.Similarity = 100.0
	entry.HashId = hex.EncodeToString(hash.Sum(nil))

	entries = append(entries, entry)

	return entries, nil
}
