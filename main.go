package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	videoURL := os.Getenv("VIDEO_URL")
	if videoURL == "" {
		fmt.Fprintln(os.Stderr, "VIDEO_URL is required")
		os.Exit(1)
	}

	// Always download the best possible quality
	formatStr := "bestvideo+bestaudio/best"

	// Build yt-dlp arguments with robust flags
	args := []string{
		"--no-playlist",
		"-f", formatStr,
		"--merge-output-format", "mp4",
		"-o", "%(title).100s [%(id)s].%(ext)s",
		"--no-mtime",
		"--retries", "10",
		"--fragment-retries", "10",
		"--concurrent-fragments", "4",
		"--no-check-certificates",
		"--prefer-free-formats",
		"--no-warnings",
	}

	// Use SOCKS5 proxy if provided (WARP bypass)
	if proxyAddr := os.Getenv("PROXY_ADDRESS"); proxyAddr != "" {
		args = append(args, "--proxy", proxyAddr)
	}

	args = append(args, videoURL)

	cmd := exec.Command("yt-dlp", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Running yt-dlp (best quality)...")
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

	// Zip the video
	zipPath := filepath.Join(downloadDir, "video.zip")
	if err := createZip(zipPath, newPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create zip: %v\n", err)
		os.Exit(1)
	}

	// Delete the original MP4 file (keep only the zip)
	os.Remove(newPath)

	fmt.Printf("Successfully created: %s\n", zipPath)
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
