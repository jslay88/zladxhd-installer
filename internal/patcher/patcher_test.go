package patcher_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jslay88/zladxhd-installer/internal/patcher"
)

var _ = Describe("Patcher", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "patcher-test-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	Describe("NewPatcher", func() {
		It("should create a patcher with correct paths", func() {
			gameDir := filepath.Join(tmpDir, "game")
			cacheDir := filepath.Join(tmpDir, "cache")

			p := patcher.NewPatcher(gameDir, cacheDir)
			Expect(p.GameDir).To(Equal(gameDir))
			Expect(p.CacheDir).To(Equal(cacheDir))
			Expect(p.PatcherPath).To(BeEmpty())
		})
	})

	Describe("FindPatcherAsset", func() {
		It("should find patcher asset with matching name", func() {
			release := &patcher.Release{
				TagName: "v1.0.0",
				Assets: []patcher.Asset{
					{Name: "readme.txt", DownloadURL: "https://example.com/readme.txt"},
					{Name: "LADXHD.Patcher.exe", DownloadURL: "https://example.com/patcher.exe"},
				},
			}

			asset, err := patcher.FindPatcherAsset(release)
			Expect(err).NotTo(HaveOccurred())
			Expect(asset.Name).To(Equal("LADXHD.Patcher.exe"))
		})

		It("should be case-insensitive for patcher name", func() {
			release := &patcher.Release{
				TagName: "v1.0.0",
				Assets: []patcher.Asset{
					{Name: "ladxhd.patcher.EXE", DownloadURL: "https://example.com/patcher.exe"},
				},
			}

			asset, err := patcher.FindPatcherAsset(release)
			Expect(err).NotTo(HaveOccurred())
			Expect(asset).NotTo(BeNil())
		})

		It("should fallback to any .exe file", func() {
			release := &patcher.Release{
				TagName: "v1.0.0",
				Assets: []patcher.Asset{
					{Name: "readme.txt", DownloadURL: "https://example.com/readme.txt"},
					{Name: "SomeOther.exe", DownloadURL: "https://example.com/other.exe"},
				},
			}

			asset, err := patcher.FindPatcherAsset(release)
			Expect(err).NotTo(HaveOccurred())
			Expect(asset.Name).To(Equal("SomeOther.exe"))
		})

		It("should return error when no executable found", func() {
			release := &patcher.Release{
				TagName: "v1.0.0",
				Assets: []patcher.Asset{
					{Name: "readme.txt", DownloadURL: "https://example.com/readme.txt"},
					{Name: "archive.zip", DownloadURL: "https://example.com/archive.zip"},
				},
			}

			_, err := patcher.FindPatcherAsset(release)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("patcher executable not found"))
		})
	})

	Describe("FindExisting", func() {
		It("should find existing patcher in game directory", func() {
			gameDir := filepath.Join(tmpDir, "game")
			err := os.MkdirAll(gameDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			// Create a fake patcher file
			patcherFile := filepath.Join(gameDir, "LADXHD.Patcher.exe")
			err = os.WriteFile(patcherFile, []byte("fake exe"), 0644)
			Expect(err).NotTo(HaveOccurred())

			p := patcher.NewPatcher(gameDir, tmpDir)
			foundPath, err := p.FindExisting()
			Expect(err).NotTo(HaveOccurred())
			Expect(foundPath).To(Equal(patcherFile))
			Expect(p.PatcherPath).To(Equal(patcherFile))
		})

		It("should be case-insensitive", func() {
			gameDir := filepath.Join(tmpDir, "game")
			err := os.MkdirAll(gameDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			// Create a fake patcher file with different case
			patcherFile := filepath.Join(gameDir, "ladxhd.PATCHER.exe")
			err = os.WriteFile(patcherFile, []byte("fake exe"), 0644)
			Expect(err).NotTo(HaveOccurred())

			p := patcher.NewPatcher(gameDir, tmpDir)
			foundPath, err := p.FindExisting()
			Expect(err).NotTo(HaveOccurred())
			Expect(foundPath).To(Equal(patcherFile))
		})

		It("should return error when patcher not found", func() {
			gameDir := filepath.Join(tmpDir, "game")
			err := os.MkdirAll(gameDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			p := patcher.NewPatcher(gameDir, tmpDir)
			_, err = p.FindExisting()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("patcher not found"))
		})

		It("should return error for non-existent game directory", func() {
			p := patcher.NewPatcher(filepath.Join(tmpDir, "nonexistent"), tmpDir)
			_, err := p.FindExisting()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetDownloadURL", func() {
		It("should construct correct download URL", func() {
			url := patcher.GetDownloadURL("v1.2.3", "LADXHD.Patcher.exe")
			Expect(url).To(Equal("https://github.com/BigheadSMZ/Zelda-LA-DX-HD-Updated/releases/download/v1.2.3/LADXHD.Patcher.exe"))
		})
	})

	Describe("Run", func() {
		It("should return error when patcher path is empty", func() {
			p := patcher.NewPatcher(tmpDir, tmpDir)
			// PatcherPath is empty by default

			err := p.Run(nil, 12345, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("patcher not downloaded"))
		})
	})

	Describe("Constants", func() {
		It("should have correct GitHub repo", func() {
			Expect(patcher.GitHubRepo).To(Equal("BigheadSMZ/Zelda-LA-DX-HD-Updated"))
		})

		It("should have correct patcher name pattern", func() {
			Expect(patcher.PatcherNamePattern).To(Equal("LADXHD.Patcher"))
		})
	})
})
