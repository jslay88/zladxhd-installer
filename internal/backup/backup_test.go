package backup_test

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jslay88/zladxhd-installer/internal/backup"
)

var _ = Describe("Backup", func() {
	var tmpDir string
	var steamDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "backup-test-*")
		Expect(err).NotTo(HaveOccurred())

		// Create a mock Steam directory
		steamDir = filepath.Join(tmpDir, "Steam")
		err = os.MkdirAll(filepath.Join(steamDir, "config"), 0755)
		Expect(err).NotTo(HaveOccurred())
		err = os.MkdirAll(filepath.Join(steamDir, "userdata", "12345"), 0755)
		Expect(err).NotTo(HaveOccurred())
		err = os.MkdirAll(filepath.Join(steamDir, "steamapps", "common"), 0755)
		Expect(err).NotTo(HaveOccurred())

		// Create some test files
		err = os.WriteFile(filepath.Join(steamDir, "config", "config.vdf"), []byte("config content"), 0644)
		Expect(err).NotTo(HaveOccurred())
		err = os.WriteFile(filepath.Join(steamDir, "userdata", "12345", "localconfig.vdf"), []byte("localconfig content"), 0644)
		Expect(err).NotTo(HaveOccurred())
		err = os.WriteFile(filepath.Join(steamDir, "steamapps", "common", "game.txt"), []byte("game file"), 0644)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	Describe("DefaultOptions", func() {
		It("should return default options with steamapps excluded", func() {
			opts := backup.DefaultOptions("/path/to/steam")
			Expect(opts.SteamPath).To(Equal("/path/to/steam"))
			Expect(opts.ExcludeDirs).To(ContainElement("steamapps"))
		})

		It("should set output dir to home directory", func() {
			opts := backup.DefaultOptions("/path/to/steam")
			home, _ := os.UserHomeDir()
			Expect(opts.OutputDir).To(Equal(home))
		})
	})

	Describe("Create", func() {
		It("should create a backup archive", func() {
			outputDir := filepath.Join(tmpDir, "output")
			err := os.MkdirAll(outputDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			opts := backup.Options{
				SteamPath:   steamDir,
				OutputDir:   outputDir,
				ExcludeDirs: []string{"steamapps"},
			}

			result, err := backup.Create(opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Path).To(ContainSubstring("Steam-backup-"))
			Expect(result.Path).To(HaveSuffix(".tar.gz"))
			Expect(result.Size).To(BeNumerically(">", 0))
			Expect(result.FileCount).To(BeNumerically(">", 0))
			Expect(result.Duration).To(BeNumerically(">", 0))
		})

		It("should exclude specified directories", func() {
			outputDir := filepath.Join(tmpDir, "output")
			err := os.MkdirAll(outputDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			opts := backup.Options{
				SteamPath:   steamDir,
				OutputDir:   outputDir,
				ExcludeDirs: []string{"steamapps"},
			}

			result, err := backup.Create(opts)
			Expect(err).NotTo(HaveOccurred())

			// Read the archive and verify steamapps is not included
			file, err := os.Open(result.Path)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = file.Close() }()

			gzReader, err := gzip.NewReader(file)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = gzReader.Close() }()

			tarReader := tar.NewReader(gzReader)
			var foundSteamapps bool
			for {
				header, err := tarReader.Next()
				if err == io.EOF {
					break
				}
				Expect(err).NotTo(HaveOccurred())
				if filepath.Base(header.Name) == "steamapps" {
					foundSteamapps = true
				}
			}
			Expect(foundSteamapps).To(BeFalse())
		})

		It("should call progress callback", func() {
			outputDir := filepath.Join(tmpDir, "output")
			err := os.MkdirAll(outputDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			var progressCalls int
			opts := backup.Options{
				SteamPath:   steamDir,
				OutputDir:   outputDir,
				ExcludeDirs: []string{"steamapps"},
				OnProgress: func(current, total int64, currentFile string) {
					progressCalls++
				},
			}

			_, err = backup.Create(opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(progressCalls).To(BeNumerically(">", 0))
		})

		It("should handle non-existent Steam path gracefully", func() {
			opts := backup.Options{
				SteamPath:   filepath.Join(tmpDir, "nonexistent"),
				OutputDir:   tmpDir,
				ExcludeDirs: []string{},
			}

			// filepath.Walk handles non-existent paths gracefully
			// The backup will be created but with 0 files
			result, err := backup.Create(opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.FileCount).To(Equal(0))
		})
	})

	Describe("FormatSize", func() {
		DescribeTable("should format sizes correctly",
			func(bytes int64, expected string) {
				Expect(backup.FormatSize(bytes)).To(Equal(expected))
			},
			Entry("bytes", int64(500), "500 B"),
			Entry("kilobytes", int64(1024), "1.0 KB"),
			Entry("megabytes", int64(1024*1024), "1.0 MB"),
			Entry("gigabytes", int64(1024*1024*1024), "1.0 GB"),
			Entry("terabytes", int64(1024*1024*1024*1024), "1.0 TB"),
			Entry("partial KB", int64(1536), "1.5 KB"),
		)
	})

	Describe("Verify", func() {
		It("should verify a valid backup", func() {
			outputDir := filepath.Join(tmpDir, "output")
			err := os.MkdirAll(outputDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			opts := backup.Options{
				SteamPath:   steamDir,
				OutputDir:   outputDir,
				ExcludeDirs: []string{"steamapps"},
			}

			result, err := backup.Create(opts)
			Expect(err).NotTo(HaveOccurred())

			err = backup.Verify(result.Path)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error for non-existent file", func() {
			err := backup.Verify(filepath.Join(tmpDir, "nonexistent.tar.gz"))
			Expect(err).To(HaveOccurred())
		})

		It("should return error for invalid gzip", func() {
			invalidFile := filepath.Join(tmpDir, "invalid.tar.gz")
			err := os.WriteFile(invalidFile, []byte("not a gzip file"), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = backup.Verify(invalidFile)
			Expect(err).To(HaveOccurred())
		})
	})
})
