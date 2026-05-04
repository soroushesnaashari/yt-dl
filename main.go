package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	videoURL := os.Getenv("VIDEO_URL")
	if videoURL == "" {
		fmt.Fprintln(os.Stderr, "VIDEO_URL is required")
		os.Exit(1)
	}

	quality := strings.ToLower(os.Getenv("QUALITY"))
	formatStr := formatSelector(quality)
	fmt.Printf("Quality: %s -> Format selector: %s\n", quality, formatStr)

	// Optional cookies file (if you ever need it)
	cookiesFile := os.Getenv("COOKIES_FILE")

	// Prepare yt-dlp arguments
	args := []string{
		"--no-playlist",
		"--extractor-args", "youtube:player_client=ios,android",
		"-f", formatStr,
		"--merge-output-format", "mp4",
		"-o", "%(title).100s [%(id)s].%(ext)s",
		"--no-mtime",
	}
	if cookiesFile != "" {
		args = append(args, "--cookies", cookiesFile)
	}
	args = append(args, videoURL)

	cmd := exec.Command("yt-dlp", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Running yt-dlp...")
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "yt-dlp failed: %v\n", err)
		os.Exit(1)
	}

	// Locate the downloaded MP4
	matches, err := filepath.Glob("*.mp4")
	if err != nil || len(matches) == 0 {
		fmt.Fprintln(os.Stderr, "No MP4 file found after download")
		os.Exit(1)
	}
	videoFile := matches[0]
	fmt.Printf("Downloaded: %s\n", videoFile)

	// Create download folder
	downloadDir := "download"
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create download dir: %v\n", err)
		os.Exit(1)
	}

	// Move video into download folder
	newPath := filepath.Join(downloadDir, filepath.Base(videoFile))
	if err := os.Rename(videoFile, newPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to move video: %v\n", err)
		os.Exit(1)
	}

	// Zip the video inside the download folder
	zipPath := filepath.Join(downloadDir, "video.zip")
	if err := createZip(zipPath, newPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create zip: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created: %s\n", zipPath)
}

// formatSelector builds a yt-dlp format string that tries the chosen
// quality first, then 1080p, then 720p, then 480p, then 360p,
// and finally falls back to the best available stream.
func formatSelector(q string) string {
	heights := []string{q, "1080", "720", "480", "360"}

	// Deduplicate while preserving order
	seen := make(map[string]bool)
	unique := make([]string, 0, len(heights))
	for _, h := range heights {
		if !seen[h] {
			seen[h] = true
			unique = append(unique, h)
		}
	}

	// Build fallback chain
	var parts []string
	for _, h := range unique {
		parts = append(parts,
			fmt.Sprintf("bestvideo[height<=%s][ext=mp4]+bestaudio[ext=m4a]", h))
	}

	// Final fallback
	finalFallback := "best[ext=mp4]/best"
	chain := append(parts, finalFallback)

	return strings.Join(chain, "/")
}

// createZip creates a zip file at zipPath containing the file sourcePath.
func createZip(zipPath, sourcePath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	srcFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	info, err := srcFile.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.Base(sourcePath)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, srcFile)
	return err
}
