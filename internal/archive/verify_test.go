package archive_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jslay88/zladxhd-installer/internal/archive"
)

var _ = Describe("Verify", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "verify-test-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	Describe("CalculateChecksum", func() {
		It("should calculate correct SHA256 checksum", func() {
			testFile := filepath.Join(tmpDir, "test.txt")
			content := []byte("hello world")
			err := os.WriteFile(testFile, content, 0644)
			Expect(err).NotTo(HaveOccurred())

			// SHA256 of "hello world"
			expectedChecksum := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"

			checksum, err := archive.CalculateChecksum(testFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(checksum).To(Equal(expectedChecksum))
		})

		It("should return error for non-existent file", func() {
			_, err := archive.CalculateChecksum(filepath.Join(tmpDir, "nonexistent.txt"))
			Expect(err).To(HaveOccurred())
		})

		It("should handle empty files", func() {
			testFile := filepath.Join(tmpDir, "empty.txt")
			err := os.WriteFile(testFile, []byte{}, 0644)
			Expect(err).NotTo(HaveOccurred())

			// SHA256 of empty string
			expectedChecksum := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

			checksum, err := archive.CalculateChecksum(testFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(checksum).To(Equal(expectedChecksum))
		})
	})

	Describe("VerifyChecksum", func() {
		It("should succeed with correct checksum", func() {
			testFile := filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(testFile, []byte("hello world"), 0644)
			Expect(err).NotTo(HaveOccurred())

			correctChecksum := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
			err = archive.VerifyChecksum(testFile, correctChecksum)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail with incorrect checksum", func() {
			testFile := filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(testFile, []byte("hello world"), 0644)
			Expect(err).NotTo(HaveOccurred())

			incorrectChecksum := "0000000000000000000000000000000000000000000000000000000000000000"
			err = archive.VerifyChecksum(testFile, incorrectChecksum)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("checksum mismatch"))
		})
	})

	Describe("FileExists", func() {
		It("should return true for existing file", func() {
			testFile := filepath.Join(tmpDir, "exists.txt")
			err := os.WriteFile(testFile, []byte("test"), 0644)
			Expect(err).NotTo(HaveOccurred())

			Expect(archive.FileExists(testFile)).To(BeTrue())
		})

		It("should return false for non-existent file", func() {
			Expect(archive.FileExists(filepath.Join(tmpDir, "nonexistent.txt"))).To(BeFalse())
		})

		It("should return true for existing directory", func() {
			Expect(archive.FileExists(tmpDir)).To(BeTrue())
		})
	})

	Describe("IsValidCache", func() {
		It("should return false for non-existent file", func() {
			Expect(archive.IsValidCache(filepath.Join(tmpDir, "nonexistent.zip"))).To(BeFalse())
		})

		It("should return false for file with wrong checksum", func() {
			testFile := filepath.Join(tmpDir, "wrong.zip")
			err := os.WriteFile(testFile, []byte("not the game archive"), 0644)
			Expect(err).NotTo(HaveOccurred())

			Expect(archive.IsValidCache(testFile)).To(BeFalse())
		})
	})

	Describe("VerifyExpectedChecksum", func() {
		It("should fail for file with wrong checksum", func() {
			testFile := filepath.Join(tmpDir, "wrong-expected.zip")
			err := os.WriteFile(testFile, []byte("not the expected game archive"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = archive.VerifyExpectedChecksum(testFile)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("checksum mismatch"))
		})
	})

	Describe("ExpectedChecksum constant", func() {
		It("should be a valid SHA256 hash", func() {
			// SHA256 hashes are 64 hex characters
			Expect(archive.ExpectedChecksum).To(HaveLen(64))
			// Should be all lowercase hex characters
			for _, c := range archive.ExpectedChecksum {
				Expect(c).To(Or(
					BeNumerically(">=", '0'),
					BeNumerically("<=", '9'),
					BeNumerically(">=", 'a'),
					BeNumerically("<=", 'f'),
				))
			}
		})
	})
})
