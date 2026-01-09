package steam_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jslay88/zladxhd-installer/internal/steam"
)

var _ = Describe("Users", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "users-test-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	Describe("User", func() {
		Describe("DisplayName", func() {
			It("should show PersonaName and AccountName when both present", func() {
				user := steam.User{
					ID:          "12345",
					AccountName: "steamuser",
					PersonaName: "Cool Gamer",
				}
				Expect(user.DisplayName()).To(Equal("Cool Gamer (steamuser)"))
			})

			It("should show AccountName when PersonaName is empty", func() {
				user := steam.User{
					ID:          "12345",
					AccountName: "steamuser",
				}
				Expect(user.DisplayName()).To(Equal("steamuser"))
			})

			It("should show User ID when both names are empty", func() {
				user := steam.User{
					ID: "12345",
				}
				Expect(user.DisplayName()).To(Equal("User 12345"))
			})
		})

		Describe("ShortcutsPath", func() {
			It("should return correct path", func() {
				user := steam.User{
					ID:         "12345",
					ConfigPath: "/home/user/.local/share/Steam/userdata/12345/config",
				}
				Expect(user.ShortcutsPath()).To(Equal("/home/user/.local/share/Steam/userdata/12345/config/shortcuts.vdf"))
			})
		})

		Describe("LocalConfigPath", func() {
			It("should return correct path", func() {
				user := steam.User{
					ID:         "12345",
					ConfigPath: "/home/user/.local/share/Steam/userdata/12345/config",
				}
				Expect(user.LocalConfigPath()).To(Equal("/home/user/.local/share/Steam/userdata/12345/config/localconfig.vdf"))
			})
		})

		Describe("HasShortcuts", func() {
			It("should return false when shortcuts.vdf doesn't exist", func() {
				configPath := filepath.Join(tmpDir, "config")
				err := os.MkdirAll(configPath, 0755)
				Expect(err).NotTo(HaveOccurred())

				user := steam.User{
					ID:         "12345",
					ConfigPath: configPath,
				}
				Expect(user.HasShortcuts()).To(BeFalse())
			})

			It("should return true when shortcuts.vdf exists", func() {
				configPath := filepath.Join(tmpDir, "config")
				err := os.MkdirAll(configPath, 0755)
				Expect(err).NotTo(HaveOccurred())

				shortcutsPath := filepath.Join(configPath, "shortcuts.vdf")
				err = os.WriteFile(shortcutsPath, []byte{}, 0644)
				Expect(err).NotTo(HaveOccurred())

				user := steam.User{
					ID:         "12345",
					ConfigPath: configPath,
				}
				Expect(user.HasShortcuts()).To(BeTrue())
			})
		})
	})

	Describe("Steam.GetUsers", func() {
		var mockSteam *steam.Steam

		BeforeEach(func() {
			steamPath := filepath.Join(tmpDir, "Steam")
			err := os.MkdirAll(filepath.Join(steamPath, "steamapps"), 0755)
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
			}
		})

		It("should return error when no users found", func() {
			_, err := mockSteam.GetUsers()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no Steam users found"))
		})

		It("should find valid users", func() {
			// Create user directory with config
			userPath := filepath.Join(mockSteam.UserDataPath, "12345678", "config")
			err := os.MkdirAll(userPath, 0755)
			Expect(err).NotTo(HaveOccurred())

			users, err := mockSteam.GetUsers()
			Expect(err).NotTo(HaveOccurred())
			Expect(users).To(HaveLen(1))
			Expect(users[0].ID).To(Equal("12345678"))
		})

		It("should skip user ID 0", func() {
			// Create user 0 (should be skipped)
			user0Path := filepath.Join(mockSteam.UserDataPath, "0", "config")
			err := os.MkdirAll(user0Path, 0755)
			Expect(err).NotTo(HaveOccurred())

			// Create valid user
			userPath := filepath.Join(mockSteam.UserDataPath, "12345678", "config")
			err = os.MkdirAll(userPath, 0755)
			Expect(err).NotTo(HaveOccurred())

			users, err := mockSteam.GetUsers()
			Expect(err).NotTo(HaveOccurred())
			Expect(users).To(HaveLen(1))
			Expect(users[0].ID).To(Equal("12345678"))
		})

		It("should skip non-numeric directories", func() {
			// Create non-numeric directory
			invalidPath := filepath.Join(mockSteam.UserDataPath, "not-a-number", "config")
			err := os.MkdirAll(invalidPath, 0755)
			Expect(err).NotTo(HaveOccurred())

			// Create valid user
			userPath := filepath.Join(mockSteam.UserDataPath, "12345678", "config")
			err = os.MkdirAll(userPath, 0755)
			Expect(err).NotTo(HaveOccurred())

			users, err := mockSteam.GetUsers()
			Expect(err).NotTo(HaveOccurred())
			Expect(users).To(HaveLen(1))
		})

		It("should skip directories without config subdirectory", func() {
			// Create user directory without config
			userPath := filepath.Join(mockSteam.UserDataPath, "11111111")
			err := os.MkdirAll(userPath, 0755)
			Expect(err).NotTo(HaveOccurred())

			// Create valid user with config
			validUserPath := filepath.Join(mockSteam.UserDataPath, "12345678", "config")
			err = os.MkdirAll(validUserPath, 0755)
			Expect(err).NotTo(HaveOccurred())

			users, err := mockSteam.GetUsers()
			Expect(err).NotTo(HaveOccurred())
			Expect(users).To(HaveLen(1))
			Expect(users[0].ID).To(Equal("12345678"))
		})
	})

	Describe("Steam.GetUserByID", func() {
		var mockSteam *steam.Steam

		BeforeEach(func() {
			steamPath := filepath.Join(tmpDir, "Steam")
			err := os.MkdirAll(filepath.Join(steamPath, "steamapps"), 0755)
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
			}

			// Create test users
			err = os.MkdirAll(filepath.Join(mockSteam.UserDataPath, "12345678", "config"), 0755)
			Expect(err).NotTo(HaveOccurred())
			err = os.MkdirAll(filepath.Join(mockSteam.UserDataPath, "87654321", "config"), 0755)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should find user by ID", func() {
			user, err := mockSteam.GetUserByID("12345678")
			Expect(err).NotTo(HaveOccurred())
			Expect(user).NotTo(BeNil())
			Expect(user.ID).To(Equal("12345678"))
		})

		It("should return error for non-existent user", func() {
			_, err := mockSteam.GetUserByID("99999999")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})
	})

	Describe("Steam.GetUsers with loginusers.vdf", func() {
		var mockSteam *steam.Steam

		BeforeEach(func() {
			steamPath := filepath.Join(tmpDir, "Steam")
			err := os.MkdirAll(filepath.Join(steamPath, "steamapps"), 0755)
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
			}
		})

		It("should handle missing loginusers.vdf gracefully", func() {
			// Create user directory without loginusers.vdf
			err := os.MkdirAll(filepath.Join(mockSteam.UserDataPath, "12345678", "config"), 0755)
			Expect(err).NotTo(HaveOccurred())

			users, err := mockSteam.GetUsers()
			Expect(err).NotTo(HaveOccurred())
			Expect(users).To(HaveLen(1))
			// Names should be empty
			Expect(users[0].AccountName).To(BeEmpty())
			Expect(users[0].PersonaName).To(BeEmpty())
		})

		It("should handle invalid loginusers.vdf gracefully", func() {
			// Create user directory
			err := os.MkdirAll(filepath.Join(mockSteam.UserDataPath, "12345678", "config"), 0755)
			Expect(err).NotTo(HaveOccurred())

			// Create invalid loginusers.vdf
			loginUsersPath := filepath.Join(mockSteam.ConfigPath, "loginusers.vdf")
			err = os.WriteFile(loginUsersPath, []byte("not valid vdf"), 0644)
			Expect(err).NotTo(HaveOccurred())

			users, err := mockSteam.GetUsers()
			Expect(err).NotTo(HaveOccurred())
			Expect(users).To(HaveLen(1))
			// Names should be empty due to parse error
			Expect(users[0].AccountName).To(BeEmpty())
		})

		It("should handle loginusers.vdf without users node gracefully", func() {
			// Create user directory
			err := os.MkdirAll(filepath.Join(mockSteam.UserDataPath, "12345678", "config"), 0755)
			Expect(err).NotTo(HaveOccurred())

			// Create loginusers.vdf without "users" root
			loginUsersContent := "\"something\"\n{\n}\n"
			loginUsersPath := filepath.Join(mockSteam.ConfigPath, "loginusers.vdf")
			err = os.WriteFile(loginUsersPath, []byte(loginUsersContent), 0644)
			Expect(err).NotTo(HaveOccurred())

			users, err := mockSteam.GetUsers()
			Expect(err).NotTo(HaveOccurred())
			Expect(users).To(HaveLen(1))
			// Names should be empty
			Expect(users[0].AccountName).To(BeEmpty())
		})
	})
})
