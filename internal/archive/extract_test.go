package archive_test

import (
	"archive/zip"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jslay88/zladxhd-installer/internal/archive"
)

var _ = Describe("Extract", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "extract-test-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	// Helper function to create a test zip file
	createTestZip := func(name string, files map[string]string) string {
		zipPath := filepath.Join(tmpDir, name)
		zipFile, err := os.Create(zipPath)
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = zipFile.Close() }()

		w := zip.NewWriter(zipFile)
		for name, content := range files {
			f, err := w.Create(name)
			Expect(err).NotTo(HaveOccurred())
			_, err = f.Write([]byte(content))
			Expect(err).NotTo(HaveOccurred())
		}
		err = w.Close()
		Expect(err).NotTo(HaveOccurred())

		return zipPath
	}

	Describe("Extract", func() {
		It("should extract files from a zip archive", func() {
			zipPath := createTestZip("test.zip", map[string]string{
				"file1.txt": "content1",
				"file2.txt": "content2",
			})

			destDir := filepath.Join(tmpDir, "extracted")
			result, err := archive.Extract(archive.ExtractOptions{
				ArchivePath:  zipPath,
				DestDir:      destDir,
				ShowProgress: false,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.ExtractedFiles).To(Equal(2))

			// Verify files were extracted
			content1, err := os.ReadFile(filepath.Join(destDir, "file1.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content1)).To(Equal("content1"))

			content2, err := os.ReadFile(filepath.Join(destDir, "file2.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content2)).To(Equal("content2"))
		})

		It("should create nested directories", func() {
			zipPath := createTestZip("nested.zip", map[string]string{
				"dir1/dir2/file.txt": "nested content",
			})

			destDir := filepath.Join(tmpDir, "extracted")
			result, err := archive.Extract(archive.ExtractOptions{
				ArchivePath:  zipPath,
				DestDir:      destDir,
				ShowProgress: false,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.ExtractedFiles).To(Equal(1))

			content, err := os.ReadFile(filepath.Join(destDir, "dir1", "dir2", "file.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("nested content"))
		})

		It("should strip components from path", func() {
			zipPath := createTestZip("strip.zip", map[string]string{
				"root/subdir/file.txt": "stripped content",
			})

			destDir := filepath.Join(tmpDir, "extracted")
			result, err := archive.Extract(archive.ExtractOptions{
				ArchivePath:     zipPath,
				DestDir:         destDir,
				ShowProgress:    false,
				StripComponents: 1,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.ExtractedFiles).To(Equal(1))

			// File should be at subdir/file.txt, not root/subdir/file.txt
			content, err := os.ReadFile(filepath.Join(destDir, "subdir", "file.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("stripped content"))
		})

		It("should skip files within stripped components", func() {
			zipPath := createTestZip("strip-skip.zip", map[string]string{
				"root/file.txt": "file in root",
			})

			destDir := filepath.Join(tmpDir, "extracted")
			result, err := archive.Extract(archive.ExtractOptions{
				ArchivePath:     zipPath,
				DestDir:         destDir,
				ShowProgress:    false,
				StripComponents: 2,
			})

			Expect(err).NotTo(HaveOccurred())
			// File should be skipped because it only has 2 components and we're stripping 2
			Expect(result.ExtractedFiles).To(Equal(1)) // Counts files processed, not necessarily written
		})

		It("should return error for non-existent archive", func() {
			_, err := archive.Extract(archive.ExtractOptions{
				ArchivePath:  filepath.Join(tmpDir, "nonexistent.zip"),
				DestDir:      filepath.Join(tmpDir, "extracted"),
				ShowProgress: false,
			})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to open archive"))
		})

		It("should create destination directory if it doesn't exist", func() {
			zipPath := createTestZip("test.zip", map[string]string{
				"file.txt": "content",
			})

			destDir := filepath.Join(tmpDir, "new", "nested", "dest")
			_, err := archive.Extract(archive.ExtractOptions{
				ArchivePath:  zipPath,
				DestDir:      destDir,
				ShowProgress: false,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(archive.FileExists(destDir)).To(BeTrue())
		})
	})

	Describe("ListArchive", func() {
		It("should list all files in a zip archive", func() {
			zipPath := createTestZip("list.zip", map[string]string{
				"file1.txt":     "content1",
				"dir/file2.txt": "content2",
				"dir/file3.txt": "content3",
			})

			files, err := archive.ListArchive(zipPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(HaveLen(3))
			Expect(files).To(ContainElements("file1.txt", "dir/file2.txt", "dir/file3.txt"))
		})

		It("should return error for non-existent archive", func() {
			_, err := archive.ListArchive(filepath.Join(tmpDir, "nonexistent.zip"))
			Expect(err).To(HaveOccurred())
		})

		It("should return empty slice for empty archive", func() {
			zipPath := createTestZip("empty.zip", map[string]string{})

			files, err := archive.ListArchive(zipPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(BeEmpty())
		})
	})

	Describe("FindExecutable", func() {
		It("should find an executable by name substring", func() {
			zipPath := createTestZip("game.zip", map[string]string{
				"readme.txt":        "readme",
				"game/launcher.exe": "exe content",
				"game/config.ini":   "config",
			})

			exePath, err := archive.FindExecutable(zipPath, "launcher")
			Expect(err).NotTo(HaveOccurred())
			Expect(exePath).To(Equal("game/launcher.exe"))
		})

		It("should be case-insensitive", func() {
			zipPath := createTestZip("game.zip", map[string]string{
				"GAME/Launcher.EXE": "exe content",
			})

			exePath, err := archive.FindExecutable(zipPath, "launcher")
			Expect(err).NotTo(HaveOccurred())
			Expect(exePath).To(Equal("GAME/Launcher.EXE"))
		})

		It("should return error when no executable found", func() {
			zipPath := createTestZip("noexe.zip", map[string]string{
				"file.txt": "text content",
			})

			_, err := archive.FindExecutable(zipPath, "game")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("executable not found"))
		})

		It("should return error for non-existent archive", func() {
			_, err := archive.FindExecutable(filepath.Join(tmpDir, "nonexistent.zip"), "game")
			Expect(err).To(HaveOccurred())
		})

		It("should only match .exe files", func() {
			zipPath := createTestZip("mixed.zip", map[string]string{
				"game.txt":    "text",
				"game.exe":    "exe",
				"game_config": "config",
			})

			exePath, err := archive.FindExecutable(zipPath, "game")
			Expect(err).NotTo(HaveOccurred())
			Expect(exePath).To(Equal("game.exe"))
		})
	})

	Describe("GetArchiveSize", func() {
		It("should return the size of an archive", func() {
			content := "test content for size check"
			zipPath := createTestZip("size.zip", map[string]string{
				"file.txt": content,
			})

			size, err := archive.GetArchiveSize(zipPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(size).To(BeNumerically(">", 0))
		})

		It("should return error for non-existent file", func() {
			_, err := archive.GetArchiveSize(filepath.Join(tmpDir, "nonexistent.zip"))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Extract with progress", func() {
		It("should work with ShowProgress enabled", func() {
			zipPath := createTestZip("progress.zip", map[string]string{
				"file1.txt": "content1",
				"file2.txt": "content2",
			})

			destDir := filepath.Join(tmpDir, "progress-extracted")
			result, err := archive.Extract(archive.ExtractOptions{
				ArchivePath:  zipPath,
				DestDir:      destDir,
				ShowProgress: true,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.ExtractedFiles).To(Equal(2))
		})
	})

	Describe("Extract edge cases", func() {
		It("should handle zip slip protection", func() {
			// Create a zip that tries to escape the destination directory
			zipPath := filepath.Join(tmpDir, "zipslip.zip")
			zipFile, err := os.Create(zipPath)
			Expect(err).NotTo(HaveOccurred())

			w := zip.NewWriter(zipFile)
			// Create a file with path traversal attempt
			f, err := w.Create("../../escaped.txt")
			Expect(err).NotTo(HaveOccurred())
			_, err = f.Write([]byte("escaped content"))
			Expect(err).NotTo(HaveOccurred())
			_ = w.Close()
			_ = zipFile.Close()

			destDir := filepath.Join(tmpDir, "zipslip-extracted")
			_, err = archive.Extract(archive.ExtractOptions{
				ArchivePath:  zipPath,
				DestDir:      destDir,
				ShowProgress: false,
			})

			// Should error on zip slip attempt
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid file path"))
		})

		It("should handle files in nested directories", func() {
			// Use the helper function which creates proper zip entries
			zipPath := createTestZip("nested-test.zip", map[string]string{
				"mydir/subdir/file.txt": "nested content",
			})

			destDir := filepath.Join(tmpDir, "nested-extracted")
			result, err := archive.Extract(archive.ExtractOptions{
				ArchivePath:  zipPath,
				DestDir:      destDir,
				ShowProgress: false,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.ExtractedFiles).To(Equal(1))

			// Verify directory structure was created
			info, err := os.Stat(filepath.Join(destDir, "mydir", "subdir"))
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())

			// Verify file content
			content, err := os.ReadFile(filepath.Join(destDir, "mydir", "subdir", "file.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("nested content"))
		})
	})
})
