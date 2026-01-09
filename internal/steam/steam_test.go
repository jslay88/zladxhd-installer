package steam_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jslay88/zladxhd-installer/internal/steam"
)

var _ = Describe("Steam", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "steam-test-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	Describe("DefaultSteamPath", func() {
		It("should be correct", func() {
			Expect(steam.DefaultSteamPath).To(Equal(".local/share/Steam"))
		})
	})

	Describe("Steam struct", func() {
		var mockSteam *steam.Steam

		BeforeEach(func() {
			// Create mock Steam directory structure
			steamPath := filepath.Join(tmpDir, "Steam")
			err := os.MkdirAll(filepath.Join(steamPath, "steamapps", "common"), 0755)
			Expect(err).NotTo(HaveOccurred())
			err = os.MkdirAll(filepath.Join(steamPath, "steamapps", "compatdata"), 0755)
			Expect(err).NotTo(HaveOccurred())
			err = os.MkdirAll(filepath.Join(steamPath, "userdata"), 0755)
			Expect(err).NotTo(HaveOccurred())
			err = os.MkdirAll(filepath.Join(steamPath, "config"), 0755)
			Expect(err).NotTo(HaveOccurred())

			mockSteam = &steam.Steam{
				Path:         steamPath,
				UserDataPath: filepath.Join(steamPath, "userdata"),
				ConfigPath:   filepath.Join(steamPath, "config"),
				AppsPath:     filepath.Join(steamPath, "steamapps"),
				CompatPath:   filepath.Join(steamPath, "steamapps", "compatdata"),
			}
		})

		Describe("CommonPath", func() {
			It("should return correct common path", func() {
				commonPath := mockSteam.CommonPath()
				Expect(commonPath).To(HaveSuffix("steamapps/common"))
			})
		})

		Describe("GetProtonVersions", func() {
			It("should return empty slice when no Proton installed", func() {
				versions, err := mockSteam.GetProtonVersions()
				Expect(err).NotTo(HaveOccurred())
				Expect(versions).To(BeEmpty())
			})

			It("should find Proton versions", func() {
				// Create some Proton directories
				protonPath := filepath.Join(mockSteam.CommonPath(), "Proton 10.0")
				err := os.MkdirAll(protonPath, 0755)
				Expect(err).NotTo(HaveOccurred())

				experimentalPath := filepath.Join(mockSteam.CommonPath(), "Proton Experimental")
				err = os.MkdirAll(experimentalPath, 0755)
				Expect(err).NotTo(HaveOccurred())

				versions, err := mockSteam.GetProtonVersions()
				Expect(err).NotTo(HaveOccurred())
				Expect(versions).To(HaveLen(2))
				Expect(versions).To(ContainElements("Proton 10.0", "Proton Experimental"))
			})

			It("should find GE-Proton versions", func() {
				gePath := filepath.Join(mockSteam.CommonPath(), "GE-Proton8-25")
				err := os.MkdirAll(gePath, 0755)
				Expect(err).NotTo(HaveOccurred())

				versions, err := mockSteam.GetProtonVersions()
				Expect(err).NotTo(HaveOccurred())
				Expect(versions).To(ContainElement("GE-Proton8-25"))
			})

			It("should ignore non-Proton directories", func() {
				gamePath := filepath.Join(mockSteam.CommonPath(), "SomeGame")
				err := os.MkdirAll(gamePath, 0755)
				Expect(err).NotTo(HaveOccurred())

				versions, err := mockSteam.GetProtonVersions()
				Expect(err).NotTo(HaveOccurred())
				Expect(versions).NotTo(ContainElement("SomeGame"))
			})
		})

		Describe("GetProtonPath", func() {
			It("should return path for existing Proton", func() {
				protonPath := filepath.Join(mockSteam.CommonPath(), "Proton 10.0")
				err := os.MkdirAll(protonPath, 0755)
				Expect(err).NotTo(HaveOccurred())

				path, err := mockSteam.GetProtonPath("Proton 10.0")
				Expect(err).NotTo(HaveOccurred())
				Expect(path).To(Equal(protonPath))
			})

			It("should return error for non-existent Proton", func() {
				_, err := mockSteam.GetProtonPath("Proton NonExistent")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})
	})

	// Note: Discover(), IsRunning(), GetPID(), Kill(), WaitForExit() are hard to test
	// They rely on actual file system paths and process management
	// These would typically be tested via integration tests
})
