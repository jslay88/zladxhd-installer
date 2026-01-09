package steam_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jslay88/zladxhd-installer/internal/steam"
)

var _ = Describe("Shortcuts", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "shortcuts-test-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	Describe("NewShortcut", func() {
		It("should create shortcut with correct defaults", func() {
			shortcut := steam.NewShortcut("Test Game", "/path/to/game.exe")

			Expect(shortcut.AppName).To(Equal("Test Game"))
			Expect(shortcut.Exe).To(Equal("\"/path/to/game.exe\""))
			Expect(shortcut.StartDir).To(Equal("\"/path/to\""))
			Expect(shortcut.AppID).To(Equal(uint32(0)))
			Expect(shortcut.IsHidden).To(Equal(uint32(0)))
			Expect(shortcut.AllowDesktopConfig).To(Equal(uint32(1)))
			Expect(shortcut.AllowOverlay).To(Equal(uint32(1)))
			Expect(shortcut.OpenVR).To(Equal(uint32(0)))
			Expect(shortcut.Devkit).To(Equal(uint32(0)))
			Expect(shortcut.Tags).NotTo(BeNil())
		})

		It("should wrap paths in quotes", func() {
			shortcut := steam.NewShortcut("Game", "/some/path/with spaces/game.exe")

			Expect(shortcut.Exe).To(ContainSubstring("/some/path/with spaces/game.exe"))
			Expect(shortcut.Exe).To(HavePrefix("\""))
			Expect(shortcut.Exe).To(HaveSuffix("\""))
		})
	})

	Describe("GenerateAppID", func() {
		It("should generate AppID in non-Steam game range", func() {
			appID, err := steam.GenerateAppID()
			Expect(err).NotTo(HaveOccurred())
			Expect(appID).To(BeNumerically(">=", uint32(0xFF000000)))
			Expect(appID).To(BeNumerically("<=", uint32(0xFFFFFFFF)))
		})

		It("should generate different AppIDs", func() {
			appID1, err := steam.GenerateAppID()
			Expect(err).NotTo(HaveOccurred())

			appID2, err := steam.GenerateAppID()
			Expect(err).NotTo(HaveOccurred())

			// Very unlikely to be the same
			Expect(appID1).NotTo(Equal(appID2))
		})
	})

	Describe("ReadShortcuts", func() {
		It("should return empty slice for non-existent file", func() {
			shortcuts, err := steam.ReadShortcuts(filepath.Join(tmpDir, "nonexistent.vdf"))
			Expect(err).NotTo(HaveOccurred())
			Expect(shortcuts).To(BeEmpty())
		})

		It("should return empty slice for empty file", func() {
			emptyFile := filepath.Join(tmpDir, "empty.vdf")
			err := os.WriteFile(emptyFile, []byte{}, 0644)
			Expect(err).NotTo(HaveOccurred())

			shortcuts, err := steam.ReadShortcuts(emptyFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(shortcuts).To(BeEmpty())
		})
	})

	Describe("WriteShortcuts and ReadShortcuts roundtrip", func() {
		It("should write and read shortcuts correctly", func() {
			shortcutPath := filepath.Join(tmpDir, "config", "shortcuts.vdf")

			original := []steam.Shortcut{
				{
					AppID:              0xFF000001,
					AppName:            "Test Game 1",
					Exe:                "\"/path/to/game1.exe\"",
					StartDir:           "\"/path/to\"",
					AllowDesktopConfig: 1,
					AllowOverlay:       1,
					Tags:               map[string]string{"0": "favorite"},
				},
				{
					AppID:              0xFF000002,
					AppName:            "Test Game 2",
					Exe:                "\"/path/to/game2.exe\"",
					StartDir:           "\"/path/to\"",
					AllowDesktopConfig: 1,
					AllowOverlay:       1,
					Tags:               map[string]string{},
				},
			}

			err := steam.WriteShortcuts(shortcutPath, original)
			Expect(err).NotTo(HaveOccurred())

			loaded, err := steam.ReadShortcuts(shortcutPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(loaded).To(HaveLen(2))

			// Find Test Game 1
			var game1 *steam.Shortcut
			for i := range loaded {
				if loaded[i].AppName == "Test Game 1" {
					game1 = &loaded[i]
					break
				}
			}
			Expect(game1).NotTo(BeNil())
			Expect(game1.AppID).To(Equal(uint32(0xFF000001)))
			Expect(game1.Exe).To(Equal("\"/path/to/game1.exe\""))
		})
	})

	Describe("AddShortcut", func() {
		var mockUser *steam.User
		var configPath string

		BeforeEach(func() {
			configPath = filepath.Join(tmpDir, "config")
			err := os.MkdirAll(configPath, 0755)
			Expect(err).NotTo(HaveOccurred())

			mockUser = &steam.User{
				ID:         "12345",
				ConfigPath: configPath,
			}
		})

		It("should add a new shortcut", func() {
			shortcut := steam.NewShortcut("New Game", "/path/to/game.exe")

			appID, isNew, err := steam.AddShortcut(mockUser, shortcut)
			Expect(err).NotTo(HaveOccurred())
			Expect(isNew).To(BeTrue())
			Expect(appID).To(BeNumerically(">=", uint32(0xFF000000)))

			// Verify it was added
			shortcuts, err := steam.ReadShortcuts(mockUser.ShortcutsPath())
			Expect(err).NotTo(HaveOccurred())
			Expect(shortcuts).To(HaveLen(1))
			Expect(shortcuts[0].AppName).To(Equal("New Game"))
		})

		It("should return existing shortcut if name matches", func() {
			// Add first shortcut
			shortcut1 := steam.NewShortcut("Existing Game", "/path/to/game.exe")
			appID1, isNew1, err := steam.AddShortcut(mockUser, shortcut1)
			Expect(err).NotTo(HaveOccurred())
			Expect(isNew1).To(BeTrue())

			// Try to add shortcut with same name
			shortcut2 := steam.NewShortcut("Existing Game", "/different/path.exe")
			appID2, isNew2, err := steam.AddShortcut(mockUser, shortcut2)
			Expect(err).NotTo(HaveOccurred())
			Expect(isNew2).To(BeFalse())
			Expect(appID2).To(Equal(appID1))

			// Should still only have one shortcut
			shortcuts, err := steam.ReadShortcuts(mockUser.ShortcutsPath())
			Expect(err).NotTo(HaveOccurred())
			Expect(shortcuts).To(HaveLen(1))
		})

		It("should handle provided AppID", func() {
			shortcut := steam.NewShortcut("Game With ID", "/path/to/game.exe")
			shortcut.AppID = 0xFF123456

			appID, isNew, err := steam.AddShortcut(mockUser, shortcut)
			Expect(err).NotTo(HaveOccurred())
			Expect(isNew).To(BeTrue())
			Expect(appID).To(Equal(uint32(0xFF123456)))
		})
	})

	Describe("UpdateShortcut", func() {
		var mockUser *steam.User
		var configPath string

		BeforeEach(func() {
			configPath = filepath.Join(tmpDir, "config")
			err := os.MkdirAll(configPath, 0755)
			Expect(err).NotTo(HaveOccurred())

			mockUser = &steam.User{
				ID:         "12345",
				ConfigPath: configPath,
			}
		})

		It("should update existing shortcut", func() {
			// Add initial shortcut
			shortcut := steam.NewShortcut("Game", "/path/to/game.exe")
			appID, _, err := steam.AddShortcut(mockUser, shortcut)
			Expect(err).NotTo(HaveOccurred())

			// Update it
			shortcut.AppID = appID
			shortcut.AppName = "Updated Game Name"
			err = steam.UpdateShortcut(mockUser, shortcut)
			Expect(err).NotTo(HaveOccurred())

			// Verify update
			shortcuts, err := steam.ReadShortcuts(mockUser.ShortcutsPath())
			Expect(err).NotTo(HaveOccurred())
			Expect(shortcuts[0].AppName).To(Equal("Updated Game Name"))
		})

		It("should return error for non-existent shortcut", func() {
			shortcut := &steam.Shortcut{
				AppID:   0xFFFFFFFF,
				AppName: "NonExistent",
			}

			err := steam.UpdateShortcut(mockUser, shortcut)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})
	})

	Describe("FindShortcutByName", func() {
		var mockUser *steam.User
		var configPath string

		BeforeEach(func() {
			configPath = filepath.Join(tmpDir, "config")
			err := os.MkdirAll(configPath, 0755)
			Expect(err).NotTo(HaveOccurred())

			mockUser = &steam.User{
				ID:         "12345",
				ConfigPath: configPath,
			}
		})

		It("should find shortcut by name", func() {
			shortcut := steam.NewShortcut("Findable Game", "/path/to/game.exe")
			_, _, err := steam.AddShortcut(mockUser, shortcut)
			Expect(err).NotTo(HaveOccurred())

			found, err := steam.FindShortcutByName(mockUser, "Findable Game")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).NotTo(BeNil())
			Expect(found.AppName).To(Equal("Findable Game"))
		})

		It("should return nil for non-existent shortcut", func() {
			found, err := steam.FindShortcutByName(mockUser, "NonExistent")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeNil())
		})
	})

	Describe("RemoveShortcut", func() {
		var mockUser *steam.User
		var configPath string

		BeforeEach(func() {
			configPath = filepath.Join(tmpDir, "config")
			err := os.MkdirAll(configPath, 0755)
			Expect(err).NotTo(HaveOccurred())

			mockUser = &steam.User{
				ID:         "12345",
				ConfigPath: configPath,
			}
		})

		It("should remove shortcut by AppID", func() {
			// Add two shortcuts
			shortcut1 := steam.NewShortcut("Game 1", "/path/to/game1.exe")
			appID1, _, err := steam.AddShortcut(mockUser, shortcut1)
			Expect(err).NotTo(HaveOccurred())

			shortcut2 := steam.NewShortcut("Game 2", "/path/to/game2.exe")
			_, _, err = steam.AddShortcut(mockUser, shortcut2)
			Expect(err).NotTo(HaveOccurred())

			// Remove the first one
			err = steam.RemoveShortcut(mockUser, appID1)
			Expect(err).NotTo(HaveOccurred())

			// Verify only one remains
			shortcuts, err := steam.ReadShortcuts(mockUser.ShortcutsPath())
			Expect(err).NotTo(HaveOccurred())
			Expect(shortcuts).To(HaveLen(1))
			Expect(shortcuts[0].AppName).To(Equal("Game 2"))
		})

		It("should handle removing non-existent shortcut gracefully", func() {
			// Write an empty shortcuts file first
			err := steam.WriteShortcuts(mockUser.ShortcutsPath(), []steam.Shortcut{})
			Expect(err).NotTo(HaveOccurred())

			// Try to remove a non-existent shortcut
			err = steam.RemoveShortcut(mockUser, 0xFFFFFFFF)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Shortcut struct", func() {
		Describe("toVDFMap", func() {
			It("should convert shortcut to VDF map with all fields", func() {
				shortcut := &steam.Shortcut{
					AppID:               0xFF000001,
					AppName:             "Test",
					Exe:                 "\"/path/to/game.exe\"",
					StartDir:            "\"/path/to\"",
					Icon:                "/path/to/icon.ico",
					ShortcutPath:        "",
					LaunchOptions:       "-windowed",
					IsHidden:            0,
					AllowDesktopConfig:  1,
					AllowOverlay:        1,
					OpenVR:              0,
					Devkit:              0,
					DevkitGameID:        "",
					DevkitOverrideAppID: 0,
					LastPlayTime:        12345,
					FlatpakAppID:        "",
					Tags:                map[string]string{"0": "tag1", "1": "tag2"},
				}

				// Write and read to test conversion
				configPath := filepath.Join(tmpDir, "toVDFMap")
				err := os.MkdirAll(configPath, 0755)
				Expect(err).NotTo(HaveOccurred())

				user := &steam.User{ID: "12345", ConfigPath: configPath}
				err = steam.WriteShortcuts(user.ShortcutsPath(), []steam.Shortcut{*shortcut})
				Expect(err).NotTo(HaveOccurred())

				loaded, err := steam.ReadShortcuts(user.ShortcutsPath())
				Expect(err).NotTo(HaveOccurred())
				Expect(loaded).To(HaveLen(1))
				Expect(loaded[0].AppID).To(Equal(uint32(0xFF000001)))
				Expect(loaded[0].LaunchOptions).To(Equal("-windowed"))
				Expect(loaded[0].Icon).To(Equal("/path/to/icon.ico"))
				Expect(loaded[0].LastPlayTime).To(Equal(uint32(12345)))
			})
		})
	})
})
