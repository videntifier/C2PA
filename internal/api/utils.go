package api

import (
	"log"
	"os"
	"os/exec"
)

func readBytesFromPlaylist(playlistUrl string) ([]byte, error) {
	log.Printf("[INFO] Downloading file from playlist")
	// Create a temporary file
	tmpFile, err := os.CreateTemp(".", "playlist-*.mp4")
	if err != nil {
		return nil, err
	}

	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Use yt-dlp to download the media file
	cmd := exec.Command("yt-dlp", "--no-continue", "--force-overwrites", "-o", tmpFile.Name(), playlistUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// Read the temporary file into a byte array
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		log.Printf("[ERROR] Failed to read file: %s", err.Error())
		return nil, err
	}

	return data, nil
}
