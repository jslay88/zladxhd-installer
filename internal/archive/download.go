// Package archive handles game archive download, verification, and extraction.
package archive

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

// IsURL checks if a string is a URL.
func IsURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// DownloadOptions configures the download operation.
type DownloadOptions struct {
	// URL is the URL to download from.
	URL string
	// DestPath is the destination file path.
	DestPath string
	// ShowProgress enables a progress bar.
	ShowProgress bool
}

// Download downloads a file from a URL.
func Download(opts DownloadOptions) error {
	// Create destination directory if needed
	dir := filepath.Dir(opts.DestPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create temporary file for download
	tmpPath := opts.DestPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = out.Close() }()

	// Start download
	resp, err := http.Get(opts.URL)
	if err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to download: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Get content length for progress bar
	var reader io.Reader = resp.Body

	if opts.ShowProgress && resp.ContentLength > 0 {
		bar := progressbar.NewOptions64(
			resp.ContentLength,
			progressbar.OptionSetDescription("Downloading"),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(40),
			progressbar.OptionThrottle(100),
			progressbar.OptionShowCount(),
			progressbar.OptionOnCompletion(func() {
				fmt.Fprint(os.Stderr, "\n")
			}),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
			progressbar.OptionSetRenderBlankState(true),
		)
		reader = io.TeeReader(resp.Body, bar)
	}

	// Copy to file
	if _, err := io.Copy(out, reader); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Close the file before rename
	_ = out.Close()

	// Rename temp file to final destination
	if err := os.Rename(tmpPath, opts.DestPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to finalize download: %w", err)
	}

	return nil
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	// Create destination directory if needed
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = sourceFile.Close() }()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = destFile.Close() }()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}
