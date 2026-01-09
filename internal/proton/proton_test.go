package proton_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jslay88/zladxhd-installer/internal/proton"
	"github.com/jslay88/zladxhd-installer/internal/steam"
)

var _ = Describe("Proton", func() {
	var tmpDir string
	var mockSteam *steam.Steam
	var mockUser *steam.User

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "proton-test-*")
		Expect(err).NotTo(HaveOccurred())

		// Create mock Steam structure
		steamPath := filepath.Join(tmpDir, "Steam")
		commonPath := filepath.Join(steamPath, "steamapps", "common")
		compatPath := filepath.Join(steamPath, "steamapps", "compatdata")
		userPath := filepath.Join(steamPath, "userdata", "12345", "config")
		configPath := filepath.Join(steamPath, "config")

		err = os.MkdirAll(commonPath, 0755)
		Expect(err).NotTo(HaveOccurred())
		err = os.MkdirAll(compatPath, 0755)
		Expect(err).NotTo(HaveOccurred())
		err = os.MkdirAll(userPath, 0755)
		Expect(err).NotTo(HaveOccurred())
		err = os.MkdirAll(configPath, 0755)
		Expect(err).NotTo(HaveOccurred())

		// Create mock Proton installations
		err = os.MkdirAll(filepath.Join(commonPath, "Proton 10.0"), 0755)
		Expect(err).NotTo(HaveOccurred())
		err = os.MkdirAll(filepath.Join(commonPath, "Proton Experimental"), 0755)
		Expect(err).NotTo(HaveOccurred())
		err = os.MkdirAll(filepath.Join(commonPath, "GE-Proton8-25"), 0755)
		Expect(err).NotTo(HaveOccurred())

		mockSteam = &steam.Steam{
			Path:         steamPath,
			UserDataPath: filepath.Join(steamPath, "userdata"),
			ConfigPath:   configPath,
			AppsPath:     filepath.Join(steamPath, "steamapps"),
			CompatPath:   compatPath,
		}

		mockUser = &steam.User{
			ID:         "12345",
			ConfigPath: userPath,
		}
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	Describe("NewConfig", func() {
		It("should create a config with correct paths", func() {
			cfg, err := proton.NewConfig(mockSteam, mockUser, 0xFF000001, "Proton 10.0")
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg).NotTo(BeNil())
			Expect(cfg.ProtonName).To(Equal("Proton 10.0"))
			Expect(cfg.AppID).To(Equal(uint32(0xFF000001)))
			Expect(cfg.ProtonPath).To(ContainSubstring("Proton 10.0"))
		})

		It("should return error for non-existent Proton version", func() {
			_, err := proton.NewConfig(mockSteam, mockUser, 0xFF000001, "Proton NonExistent")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Config", func() {
		var cfg *proton.Config

		BeforeEach(func() {
			var err error
			cfg, err = proton.NewConfig(mockSteam, mockUser, 0xFF000001, "Proton 10.0")
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("CompatDataPath", func() {
			It("should return correct compatdata path", func() {
				path := cfg.CompatDataPath()
				Expect(path).To(ContainSubstring("compatdata"))
				Expect(path).To(ContainSubstring("4278190081")) // 0xFF000001 in decimal
			})
		})

		Describe("PrefixPath", func() {
			It("should return correct prefix path", func() {
				path := cfg.PrefixPath()
				Expect(path).To(ContainSubstring("pfx"))
			})
		})

		Describe("CreateCompatData", func() {
			It("should create compatdata directory", func() {
				err := cfg.CreateCompatData()
				Expect(err).NotTo(HaveOccurred())

				info, err := os.Stat(cfg.CompatDataPath())
				Expect(err).NotTo(HaveOccurred())
				Expect(info.IsDir()).To(BeTrue())
			})
		})

		Describe("HasCompatData", func() {
			It("should return false when directory doesn't exist", func() {
				Expect(cfg.HasCompatData()).To(BeFalse())
			})

			It("should return true when directory exists", func() {
				err := cfg.CreateCompatData()
				Expect(err).NotTo(HaveOccurred())

				Expect(cfg.HasCompatData()).To(BeTrue())
			})
		})

		Describe("HasPrefix", func() {
			It("should return false when prefix doesn't exist", func() {
				Expect(cfg.HasPrefix()).To(BeFalse())
			})

			It("should return true when system.reg exists", func() {
				err := os.MkdirAll(cfg.PrefixPath(), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = os.WriteFile(filepath.Join(cfg.PrefixPath(), "system.reg"), []byte(""), 0644)
				Expect(err).NotTo(HaveOccurred())

				Expect(cfg.HasPrefix()).To(BeTrue())
			})
		})

		Describe("GetCompatToolName", func() {
			DescribeTable("should convert display name to internal name",
				func(protonName, expectedName string) {
					testCfg := &proton.Config{
						Steam:      mockSteam,
						ProtonName: protonName,
					}
					Expect(testCfg.GetCompatToolName()).To(Equal(expectedName))
				},
				Entry("Proton Experimental", "Proton Experimental", "proton_experimental"),
				Entry("Proton 10.0", "Proton 10.0", "proton_10"),
				Entry("Proton Hotfix", "Proton Hotfix", "proton_hotfix"),
				Entry("Proton 9.0", "Proton 9.0", "proton_9"),
			)

			It("should use folder name for custom Proton", func() {
				// Create custom Proton directory
				customPath := filepath.Join(mockSteam.Path, "compatibilitytools.d", "GE-Proton8-25")
				err := os.MkdirAll(customPath, 0755)
				Expect(err).NotTo(HaveOccurred())

				testCfg := &proton.Config{
					Steam:      mockSteam,
					ProtonName: "GE-Proton8-25",
				}
				Expect(testCfg.GetCompatToolName()).To(Equal("GE-Proton8-25"))
			})
		})
	})

	Describe("GetAvailableProtonVersions", func() {
		It("should return available Proton versions sorted correctly", func() {
			versions, err := proton.GetAvailableProtonVersions(mockSteam)
			Expect(err).NotTo(HaveOccurred())
			Expect(versions).To(HaveLen(3))
			// Experimental should be first
			Expect(versions[0]).To(Equal("Proton Experimental"))
			// GE-Proton should be second
			Expect(versions[1]).To(Equal("GE-Proton8-25"))
			// Regular Proton should be last
			Expect(versions[2]).To(Equal("Proton 10.0"))
		})
	})

	Describe("FindProtonByName", func() {
		It("should find Proton by exact name", func() {
			version, err := proton.FindProtonByName(mockSteam, "Proton 10.0")
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("Proton 10.0"))
		})

		It("should find Proton case-insensitively", func() {
			version, err := proton.FindProtonByName(mockSteam, "proton experimental")
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("Proton Experimental"))
		})

		It("should find Proton by partial match", func() {
			version, err := proton.FindProtonByName(mockSteam, "GE-Proton")
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("GE-Proton8-25"))
		})

		It("should return error when not found", func() {
			_, err := proton.FindProtonByName(mockSteam, "NonExistentProton")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})

		It("should find by normalized name (handles dashes)", func() {
			// Create a "Proton - Experimental" directory to test normalization
			protonPath := filepath.Join(mockSteam.CommonPath(), "Proton - Experimental")
			err := os.MkdirAll(protonPath, 0755)
			Expect(err).NotTo(HaveOccurred())

			// Should find it even with different dash format
			version, err := proton.FindProtonByName(mockSteam, "proton experimental")
			Expect(err).NotTo(HaveOccurred())
			// May match either "Proton Experimental" or "Proton - Experimental"
			Expect(version).To(Or(Equal("Proton Experimental"), Equal("Proton - Experimental")))
		})
	})

	Describe("Config error handling", func() {
		It("should handle CreateCompatData error for read-only directory", func() {
			// Skip this test if running as root
			if os.Getuid() == 0 {
				Skip("Cannot test permissions as root")
			}

			cfg, err := proton.NewConfig(mockSteam, mockUser, 0xFF000099, "Proton 10.0")
			Expect(err).NotTo(HaveOccurred())

			// Make compatdata read-only
			err = os.Chmod(mockSteam.CompatPath, 0444)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.Chmod(mockSteam.CompatPath, 0755) }()

			err = cfg.CreateCompatData()
			Expect(err).To(HaveOccurred())
		})
	})
})
