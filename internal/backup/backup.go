// Package backup provides Steam directory backup functionality.
package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Options configures the backup operation.
type Options struct {
	// SteamPath is the path to the Steam directory.
	SteamPath string
	// OutputDir is the directory where the backup will be saved.
	// Defaults to user's home directory.
	OutputDir string
	// ExcludeDirs is a list of directories to exclude from the backup.
	// Defaults to ["steamapps"].
	ExcludeDirs []string
	// OnProgress is called with progress updates.
	OnProgress func(current, total int64, currentFile string)
}

// DefaultOptions returns the default backup options.
func DefaultOptions(steamPath string) Options {
	home, _ := os.UserHomeDir()
	return Options{
		SteamPath:   steamPath,
		OutputDir:   home,
		ExcludeDirs: []string{"steamapps"},
	}
}

// Result contains information about the backup.
type Result struct {
	// Path is the path to the created backup file.
	Path string
	// Size is the size of the backup file in bytes.
	Size int64
	// FileCount is the number of files included in the backup.
	FileCount int
	// Duration is how long the backup took.
	Duration time.Duration
}

// Create creates a backup of the Steam directory.
func Create(opts Options) (*Result, error) {
	start := time.Now()

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("Steam-backup-%s.tar.gz", timestamp)
	backupPath := filepath.Join(opts.OutputDir, filename)

	// Create the backup file
	file, err := os.Create(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Create gzip writer
	gzWriter := gzip.NewWriter(file)
	defer func() { _ = gzWriter.Close() }()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer func() { _ = tarWriter.Close() }()

	// Build exclusion set for fast lookup
	excludeSet := make(map[string]bool)
	for _, dir := range opts.ExcludeDirs {
		excludeSet[strings.ToLower(dir)] = true
	}

	// Count files first for progress reporting
	var totalSize int64
	if opts.OnProgress != nil {
		_ = filepath.Walk(opts.SteamPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				totalSize += info.Size()
			}
			return nil
		})
	}

	// Walk the directory and add files to tar
	var currentSize int64
	var fileCount int

	err = filepath.Walk(opts.SteamPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip files we can't access
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(opts.SteamPath, path)
		if err != nil {
			return nil
		}

		// Check if this path should be excluded
		pathParts := strings.Split(relPath, string(filepath.Separator))
		for _, part := range pathParts {
			if excludeSet[strings.ToLower(part)] {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return nil
		}

		// Use relative path in tar
		header.Name = filepath.Join("Steam", relPath)

		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return nil
			}
			header.Linkname = link
			header.Typeflag = tar.TypeSymlink
		}

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		// Write file content (if it's a regular file)
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer func() { _ = file.Close() }()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return nil
			}

			fileCount++
			currentSize += info.Size()

			if opts.OnProgress != nil {
				opts.OnProgress(currentSize, totalSize, relPath)
			}
		}

		return nil
	})

	if err != nil {
		// Clean up partial backup on error
		_ = os.Remove(backupPath)
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	// Get final file size
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat backup file: %w", err)
	}

	return &Result{
		Path:      backupPath,
		Size:      fileInfo.Size(),
		FileCount: fileCount,
		Duration:  time.Since(start),
	}, nil
}

// FormatSize returns a human-readable file size.
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Verify checks if a backup file is valid.
func Verify(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer func() { _ = file.Close() }()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to read gzip: %w", err)
	}
	defer func() { _ = gzReader.Close() }()

	tarReader := tar.NewReader(gzReader)
	for {
		_, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("invalid tar archive: %w", err)
		}
	}

	return nil
}
