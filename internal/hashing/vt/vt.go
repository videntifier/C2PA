package vt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mediaguard/internal/hashing"
	"mediaguard/internal/models"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

const VSEAddress = "http://vse:7771"
const VSEToken = "VIDENTIFIER"
const VTAlgorithm = "vt"

type VTHash struct{}
type VTConfig struct{}

type VTParser struct {
	FPS string `json:"fps"`
}

// API Structs
type APIInsert struct {
	Data struct {
		ContentID int `json:"content_id"`
		Details   struct {
			Dimension         string  `json:"dimension"`
			Fps               float64 `json:"fps"`
			FpsWhenExtracting int     `json:"fps_when_extracting"`
			NrFrames          int     `json:"nr_frames"`
		} `json:"details"`
		Duration          int     `json:"duration"`
		FileSize          int     `json:"file_size"`
		Fps               float64 `json:"fps"`
		FpsWhenExtracting int     `json:"fps_when_extracting"`
		Height            int     `json:"height"`
		Md5Sum            string  `json:"md5sum"`
		NrDescs           int     `json:"nr_descs"`
		Path              string  `json:"path"`
		Scenes            []struct {
			Locations string `json:"locations"`
			SceneID   int    `json:"scene_id"`
		} `json:"scenes"`
		Sha1Sum string `json:"sha1sum"`
		Type    string `json:"type"`
		Width   int    `json:"width"`
	} `json:"data"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

type APIQueryResponse struct {
	Data struct {
		Details struct {
			NrDescsQueried     int    `json:"nr_descs_queried"`
			NrSubscenesQueried int    `json:"nr_subscenes_queried"`
			TimeFetchDescs     string `json:"time_fetch_descs"`
			TimeInitMatching   string `json:"time_init_matching"`
			TimeMatching       string `json:"time_matching"`
			TimeQuery          string `json:"time_query"`
			TimeTotalMs        string `json:"time_total_ms"`
		} `json:"details"`
		Matches []struct {
			ContentID int       `json:"content_id"`
			Coverage  string    `json:"coverage"`
			CreatedAt time.Time `json:"created_at"`
			Duration  int       `json:"duration"`
			FileSize  int       `json:"file_size"`
			Height    int       `json:"height"`
			Locations []struct {
				MatchLocations []string `json:"match_locations"`
				MatchPerc      string   `json:"match_perc"`
				QueryLocations []string `json:"query_locations"`
				QueryPerc      string   `json:"query_perc"`
				SceneID        int      `json:"scene_id"`
				VisualCopy     bool     `json:"visual_copy"`
			} `json:"locations"`
			NrDescs         int    `json:"nr_descs"`
			NrMatchingDescs int    `json:"nr_matching_descs"`
			Path            string `json:"path"`
			Sha1Sum         string `json:"sha1sum"`
			SignalStrength  int    `json:"signal_strength"`
			Type            string `json:"type"`
			Width           int    `json:"width"`
			Metadata        struct {
				CaseID   string `json:"case_id"`
				CaseType string `json:"case_type"`
			} `json:"metadata,omitempty"`
		} `json:"matches"`
		Type string `json:"type"`
	} `json:"data"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

type APIError struct {
	Message string `json:"message"`
	Code    int    `json:"error_code"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

func init() {
	hashing.Register(VTAlgorithm, &VTHash{})
}

func (h *VTHash) Name() string {
	return VTAlgorithm
}

func (h *VTHash) Description() string {
	return "Descriptor Based Hashing algorithm for highly accurate content based hashes"
}

func (h *VTHash) ExtractHash(reader io.Reader) (string, error) {

	log.Printf("[INFO] Extracting VT-hash")
	descFile, err := extractDescriptorsIntoFile(reader)

	if err != nil {
		return "", fmt.Errorf("failed to extract hash")
	}

	log.Printf("[INFO] Inserting VT Hash")
	contentId, err := insertVTHashFile(descFile)

	if err != nil {
		return "", fmt.Errorf("failed to insert hash into db")
	}

	return fmt.Sprintf("%d", contentId), nil
}

func (h *VTHash) CheckHash(reader io.Reader) ([]models.EntrySimilarity, error) {

	var entries []models.EntrySimilarity
	log.Printf("[INFO] Extracting VT-hash")

	descFile, err := extractDescriptorsIntoFile(reader)

	if err != nil {
		return entries, fmt.Errorf("failed to extract hash")
	}

	//Extract the descriptor from the file
	entries, err = queryVTHashFile(descFile)

	if err != nil {
		return entries, fmt.Errorf("failed to query hash")
	}

	return entries, nil
}

func extractDescriptorsIntoFile(reader io.Reader) (string, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("/tmp", "vt_input_*")
	if err != nil {
		log.Printf("[DEBUG] Failed to create temp file")
		return "", err
	}

	log.Printf("[DEBUG] Temp file %s", tmpFile.Name())
	defer os.Remove(tmpFile.Name())

	// Copy the reader contents to the temp file
	bytesWrt, err := io.Copy(tmpFile, reader)
	if err != nil {
		log.Printf("[DEBUG] Failed to copy contents")
		return "", err
	}

	log.Printf("[INFO] Bytes written: %d", bytesWrt)

	// Ensure all data is flushed to disk
	if err := tmpFile.Close(); err != nil {
		log.Printf("[DEBUG] Failed to close temp file")
		return "", err
	}

	// Run desc_tools to extract visual fingerprints
	descFile := tmpFile.Name() + ".desc72"
	cmd := exec.Command("./desc_tools", "--preset=optimized", tmpFile.Name(), descFile)
	if err := cmd.Run(); err != nil {
		log.Printf("[Failed to run desc_tools] %s", err.Error())
		return "", err
	}

	return descFile, nil
}

// Utility function to insert VT Fingerprints into a VT Database instance
func insertVTHashFile(filePath string) (int, error) {
	apiUrl := fmt.Sprintf("%s/api/v0/insert", VSEAddress)

	fmt.Printf("[INFO] API URL: %s \n", apiUrl)

	//fmt.Printf("[INFO] Inserting %s \n", filePath)
	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Failed to open file: " + err.Error())
		return 0, err
	}
	defer file.Close()

	// Create a new multipart form and add the file as a form argument
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))

	if err != nil {
		fmt.Println("Error creating form file:", err.Error())
		return 0, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Println("Error copying file:", err.Error())
		return 0, err
	}

	// Close the multipart writer and set the content type header
	err = writer.Close()
	if err != nil {
		return 0, err
	}

	// Create a new HTTP POST request
	postUrl, _ := url.Parse(apiUrl)
	request, err := http.NewRequest("POST", postUrl.String(), body)
	if err != nil {
		fmt.Println("Failed to create a new post request")
		return 0, err
	}

	// Tell the server to close the connection after the request
	request.Close = true

	fmt.Printf("[INFO] Inserting %s \n", filePath)
	fmt.Printf("[INFO] Using token %s \n", VSEToken)

	// Set headers for POST request
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", VSEToken)

	// Send the request and check the response status code
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("Failed to send request %s \n", err.Error())
		return 0, err
	}

	defer response.Body.Close()

	var nvtreeResponse APIInsert
	if response.StatusCode != http.StatusOK {

		fmt.Printf("Request failed %d \n", response.StatusCode)
		fmt.Printf("Request failed %s \n", response.Status)

		var vseError APIError
		err = json.NewDecoder(response.Body).Decode(&vseError)
		if err != nil {
			fmt.Printf("Failed to parse request %s \n", err.Error())
		}

		//208 Indicates that the media is already in the database
		if vseError.Code == 208 {
			log.Printf("Error msg: %s", vseError.Message)
			contentId, err := strconv.Atoi(vseError.Message)
			if err != nil {
				fmt.Printf("Failed to convert contentId: %s \n", err.Error())
				return 0, fmt.Errorf("failed to convert contentId: %v", err)
			}
			log.Printf("Returning content-id: %d", contentId)
			return contentId, nil

		} else {
			fmt.Printf("Unexpected response status: %s \n", err.Error())
			return 0, fmt.Errorf("unexpected response status code %d (%s) ", response.StatusCode, vseError.Message)
		}
	}

	err = json.NewDecoder(response.Body).Decode(&nvtreeResponse)
	if err != nil {
		fmt.Println("Failed to decode")
		return 0, err
	}

	fmt.Printf("[INFO] Inserted %s : id %d : nr_descs: %d \n", nvtreeResponse.Data.Path, nvtreeResponse.Data.ContentID, nvtreeResponse.Data.NrDescs)

	return nvtreeResponse.Data.ContentID, nil
}

func queryVTHashFile(filePath string) ([]models.EntrySimilarity, error) {
	apiUrl := fmt.Sprintf("%s/api/v0/query", VSEAddress)

	var result []models.EntrySimilarity

	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Failed to open file: " + err.Error())
		return result, err
	}
	defer file.Close()

	// Create a new multipart form and add the file as a form argument
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		fmt.Println("Error creating form file:", err.Error())
		return result, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Println("Error copying file:", err.Error())
		return result, err
	}

	err = writer.WriteField("include_metadata", "true")
	if err != nil {
		fmt.Println("Error writing field:", err.Error())
		return result, err
	}

	// Close the multipart writer and set the content type header
	err = writer.Close()
	if err != nil {
		return result, err
	}

	// Create a new HTTP POST request
	postUrl, _ := url.Parse(apiUrl)
	request, err := http.NewRequest("POST", postUrl.String(), body)
	if err != nil {
		fmt.Println("Failed to create a new post request")
		return result, err
	}

	// Tell the server to close the connection after the request
	request.Close = true

	// Set headers for POST request
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", VSEToken)

	// Send the request and check the response status code

	//start := time.Now() // Record the start time

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("Failed to send request %s \n", err.Error())
		return result, err
	}

	defer response.Body.Close()

	//duration := time.Since(start) // Calculate the elapsed time
	//fmt.Printf("[INFO] VSE Request time: %d : ms\n", duration.Milliseconds())

	if response.StatusCode != http.StatusOK {

		var vseError APIError
		err = json.NewDecoder(response.Body).Decode(&vseError)
		if err != nil {
			fmt.Printf("Failed to parse request %s \n", err.Error())
		}
		return result, fmt.Errorf("unexpected response status code %d (%s) ", response.StatusCode, vseError.Message)
	}

	var nvtreeResponse APIQueryResponse
	err = json.NewDecoder(response.Body).Decode(&nvtreeResponse)
	if err != nil {
		fmt.Println("Failed to decode")
		return result, err
	}

	//Create a similarity entry for each match
	for _, match := range nvtreeResponse.Data.Matches {

		//We use the avg query strength value * coverage for the similary signal
		queryStrengthSum := 0.0

		//Compute average queryStreingths
		for _, location := range match.Locations {
			queryPerc, err1 := strconv.ParseFloat(location.QueryPerc, 64)
			log.Printf("[DEBUG] Match Perc: %0.2f", queryPerc)
			if err1 == nil {
				queryStrengthSum += queryPerc
			}
		}

		similarityScore := queryStrengthSum / float64(len(match.Locations))

		log.Printf("[DEBUG] Similiarity score %.2f", similarityScore)

		coverage, err := strconv.ParseFloat(match.Coverage, 64)
		if err != nil {
			coverage = 0.0
		}

		log.Printf("[DEBUG] Coverage %.2f", coverage)

		if coverage > 0.0 {
			similarityScore = similarityScore * (coverage / 100.0)
		}

		var entry models.EntrySimilarity
		entry.Algorithm = VTAlgorithm
		entry.Similarity = similarityScore
		entry.HashId = fmt.Sprintf("%d", match.ContentID)

		result = append(result, entry)

	}

	return result, nil
}
