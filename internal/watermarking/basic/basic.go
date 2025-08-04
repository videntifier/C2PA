package basic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mediaguard/internal/watermarking"
	"os"
	"os/exec"
)

// LSB is a placeholder implementation of LSB watermarking.
type Basic struct {
	Algorithm string
}

func init() {

	basic := &Basic{
		Algorithm: "basic",
	}

	//If needed read configuration values from environment

	watermarking.Register(basic.Algorithm, basic)
}

// Name returns the algorithm's name.
func (w *Basic) Name() string {
	return w.Algorithm
}

// Description returns the algorithm's description.
func (w *Basic) Description() string {
	return "Basic example that embeds metadata into video"
}

// Embed returns the original media stream as a placeholder.
func (w *Basic) Embed(reader io.Reader, data []byte) (io.Reader, error) {
	// Write input to a temporary file

	inputFile, err := os.CreateTemp("", "input-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(inputFile.Name())
	_, err = io.Copy(inputFile, reader)
	if err != nil {
		inputFile.Close()
		return nil, err
	}
	inputFile.Close()

	// Prepare output file
	outputFile, err := os.CreateTemp("", "output-*.mp4")
	if err != nil {
		return nil, err
	}
	outputFile.Close()
	defer os.Remove(outputFile.Name())

	// Use ffmpeg to embed metadata
	metadata := string(data)

	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-i", inputFile.Name(),
		"-metadata", "comment="+metadata,
		"-codec", "copy",
		outputFile.Name(),
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg error: %v, %s", err, stderr.String())
	}

	// Return output as io.Reader
	out, err := os.Open(outputFile.Name())
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Extract returns a placeholder extracted data.
func (w *Basic) Extract(reader io.Reader) ([]byte, error) {
	// Write input to a temporary file
	inputFile, err := os.CreateTemp("", "input-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(inputFile.Name())
	_, err = io.Copy(inputFile, reader)
	if err != nil {
		inputFile.Close()
		return nil, err
	}
	inputFile.Close()

	// Use ffprobe to extract metadata
	cmd := exec.Command(
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		inputFile.Name(),
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffprobe error: %v", err)
	}

	// Parse JSON output to get the embedded metadata
	type tags struct {
		Format struct {
			Tags map[string]string `json:"tags"`
		} `json:"format"`
	}
	var result tags
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("json parse error: %v", err)
	}

	val, ok := result.Format.Tags["comment"]
	if !ok {
		return nil, fmt.Errorf("metadata %q not found", w.Algorithm)
	}

	return []byte(val), nil
}
