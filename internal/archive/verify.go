package archive

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// ExpectedChecksum is the SHA256 checksum of the expected game archive.
const ExpectedChecksum = "118a4adfa782b4c0097867609cb79474abaf9a95b3f684b04715a46d424beb1c"

// VerifyChecksum verifies a file's SHA256 checksum.
func VerifyChecksum(path string, expected string) error {
	actual, err := CalculateChecksum(path)
	if err != nil {
		return err
	}

	if actual != expected {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}

	return nil
}

// CalculateChecksum calculates the SHA256 checksum of a file.
func CalculateChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyExpectedChecksum verifies a file against the expected game archive checksum.
func VerifyExpectedChecksum(path string) error {
	return VerifyChecksum(path, ExpectedChecksum)
}

// FileExists checks if a file exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsValidCache checks if the cached archive exists and has the correct checksum.
func IsValidCache(cachePath string) bool {
	if !FileExists(cachePath) {
		return false
	}

	if err := VerifyExpectedChecksum(cachePath); err != nil {
		return false
	}

	return true
}
