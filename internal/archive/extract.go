package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

// ExtractOptions configures the extraction operation.
type ExtractOptions struct {
	// ArchivePath is the path to the archive file.
	ArchivePath string
	// DestDir is the destination directory.
	DestDir string
	// ShowProgress enables a progress bar.
	ShowProgress bool
	// StripComponents removes leading path components (like tar --strip-components).
	StripComponents int
}

// ExtractResult contains information about the extraction.
type ExtractResult struct {
	// ExtractedFiles is the number of files extracted.
	ExtractedFiles int
	// TotalSize is the total size of extracted files.
	TotalSize int64
}

// Extract extracts a zip archive.
func Extract(opts ExtractOptions) (*ExtractResult, error) {
	// Open the zip file
	r, err := zip.OpenReader(opts.ArchivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer func() { _ = r.Close() }()

	// Create destination directory
	if err := os.MkdirAll(opts.DestDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Calculate total size for progress bar
	var totalSize int64
	for _, f := range r.File {
		totalSize += int64(f.UncompressedSize64)
	}

	var bar *progressbar.ProgressBar
	if opts.ShowProgress && totalSize > 0 {
		bar = progressbar.NewOptions64(
			totalSize,
			progressbar.OptionSetDescription("Extracting"),
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
	}

	result := &ExtractResult{}

	for _, f := range r.File {
		if err := extractFile(f, opts.DestDir, opts.StripComponents, bar); err != nil {
			return nil, err
		}
		result.ExtractedFiles++
		result.TotalSize += int64(f.UncompressedSize64)
	}

	return result, nil
}

func extractFile(f *zip.File, destDir string, stripComponents int, bar *progressbar.ProgressBar) error {
	// Get the file path, stripping components if needed
	path := f.Name
	if stripComponents > 0 {
		parts := strings.Split(path, "/")
		if len(parts) <= stripComponents {
			// Skip this file (it's in the stripped components)
			return nil
		}
		path = filepath.Join(parts[stripComponents:]...)
	}

	// Sanitize the path to prevent zip slip
	path = filepath.Clean(path)
	if strings.HasPrefix(path, "..") {
		return fmt.Errorf("invalid file path: %s", f.Name)
	}

	destPath := filepath.Join(destDir, path)

	// Handle directories
	if f.FileInfo().IsDir() {
		return os.MkdirAll(destPath, f.Mode())
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Extract file
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in archive: %w", err)
	}
	defer func() { _ = rc.Close() }()

	outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	var writer io.Writer = outFile
	if bar != nil {
		writer = io.MultiWriter(outFile, bar)
	}

	if _, err := io.Copy(writer, rc); err != nil {
		return fmt.Errorf("failed to extract file: %w", err)
	}

	return nil
}

// ListArchive lists files in a zip archive.
func ListArchive(archivePath string) ([]string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer func() { _ = r.Close() }()

	var files []string
	for _, f := range r.File {
		files = append(files, f.Name)
	}

	return files, nil
}

// FindExecutable finds an executable file in the archive by name substring.
func FindExecutable(archivePath string, nameSubstr string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer func() { _ = r.Close() }()

	nameSubstr = strings.ToLower(nameSubstr)

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		name := strings.ToLower(filepath.Base(f.Name))
		if strings.Contains(name, nameSubstr) && strings.HasSuffix(name, ".exe") {
			return f.Name, nil
		}
	}

	return "", fmt.Errorf("executable not found: %s", nameSubstr)
}

// GetArchiveSize returns the size of an archive file.
func GetArchiveSize(archivePath string) (int64, error) {
	info, err := os.Stat(archivePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
