package archive_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jslay88/zladxhd-installer/internal/archive"
)

var _ = Describe("Download", func() {
	Describe("IsURL", func() {
		DescribeTable("should correctly identify URLs",
			func(input string, expected bool) {
				Expect(archive.IsURL(input)).To(Equal(expected))
			},
			Entry("http URL", "http://example.com/file.zip", true),
			Entry("https URL", "https://example.com/file.zip", true),
			Entry("absolute path", "/path/to/file.zip", false),
			Entry("relative path", "./file.zip", false),
			Entry("home path", "~/Downloads/file.zip", false),
			Entry("ftp URL", "ftp://example.com/file.zip", false),
		)
	})

	Describe("CopyFile", func() {
		var tmpDir string

		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "archive-test-*")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			_ = os.RemoveAll(tmpDir)
		})

		It("should copy file contents correctly", func() {
			srcFile := filepath.Join(tmpDir, "src.txt")
			dstFile := filepath.Join(tmpDir, "dst.txt")
			content := []byte("test content")

			err := os.WriteFile(srcFile, content, 0644)
			Expect(err).NotTo(HaveOccurred())

			err = archive.CopyFile(srcFile, dstFile)
			Expect(err).NotTo(HaveOccurred())

			copied, err := os.ReadFile(dstFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(copied).To(Equal(content))
		})

		It("should create destination directory if needed", func() {
			srcFile := filepath.Join(tmpDir, "src.txt")
			dstFile := filepath.Join(tmpDir, "subdir", "deep", "dst.txt")

			err := os.WriteFile(srcFile, []byte("test"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = archive.CopyFile(srcFile, dstFile)
			Expect(err).NotTo(HaveOccurred())

			Expect(archive.FileExists(dstFile)).To(BeTrue())
		})

		It("should return error for non-existent source", func() {
			srcFile := filepath.Join(tmpDir, "nonexistent.txt")
			dstFile := filepath.Join(tmpDir, "dst.txt")

			err := archive.CopyFile(srcFile, dstFile)
			Expect(err).To(HaveOccurred())
		})
	})
})
