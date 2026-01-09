package steam

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jslay88/vdf"
)

// User represents a Steam user account.
type User struct {
	ID          string // Steam3 Account ID (numeric string)
	AccountName string
	PersonaName string
	ConfigPath  string // Path to user's config directory
}

// GetUsers returns all Steam users found in userdata directory.
// Skips user ID "0" as it's typically from non-legitimate sources.
func (s *Steam) GetUsers() ([]User, error) {
	entries, err := os.ReadDir(s.UserDataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read userdata directory: %w", err)
	}

	// Try to get account names from loginusers.vdf
	accountNames := s.getLoginUsers()

	var users []User
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		id := entry.Name()

		// Skip user ID "0"
		if id == "0" {
			continue
		}

		// Verify it's a numeric ID
		if _, err := strconv.ParseUint(id, 10, 64); err != nil {
			continue
		}

		configPath := filepath.Join(s.UserDataPath, id, "config")
		if info, err := os.Stat(configPath); err != nil || !info.IsDir() {
			continue
		}

		user := User{
			ID:         id,
			ConfigPath: configPath,
		}

		// Look up account name if available
		if info, ok := accountNames[id]; ok {
			user.AccountName = info.AccountName
			user.PersonaName = info.PersonaName
		}

		users = append(users, user)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no Steam users found")
	}

	return users, nil
}

// loginUserInfo holds account information from loginusers.vdf
type loginUserInfo struct {
	AccountName string
	PersonaName string
}

// getLoginUsers reads loginusers.vdf to get account names.
func (s *Steam) getLoginUsers() map[string]loginUserInfo {
	result := make(map[string]loginUserInfo)

	loginUsersPath := filepath.Join(s.ConfigPath, "loginusers.vdf")
	doc, err := vdf.ParseFile(loginUsersPath)
	if err != nil {
		return result
	}

	users := doc.Get("users")
	if users == nil || !users.IsObject {
		return result
	}

	for _, child := range users.Children {
		if !child.IsObject {
			continue
		}

		// The key is the SteamID64, we need to convert to Steam3 ID
		steamID64, err := strconv.ParseUint(child.Key, 10, 64)
		if err != nil {
			continue
		}

		// Convert SteamID64 to Steam3 Account ID
		// Steam3 ID = SteamID64 - 76561197960265728
		steam3ID := steamID64 - 76561197960265728

		info := loginUserInfo{
			AccountName: child.GetString("AccountName"),
			PersonaName: child.GetString("PersonaName"),
		}

		result[strconv.FormatUint(steam3ID, 10)] = info
	}

	return result
}

// DisplayName returns a human-readable name for the user.
func (u *User) DisplayName() string {
	if u.PersonaName != "" {
		return fmt.Sprintf("%s (%s)", u.PersonaName, u.AccountName)
	}
	if u.AccountName != "" {
		return u.AccountName
	}
	return fmt.Sprintf("User %s", u.ID)
}

// ShortcutsPath returns the path to the user's shortcuts.vdf file.
func (u *User) ShortcutsPath() string {
	return filepath.Join(u.ConfigPath, "shortcuts.vdf")
}

// LocalConfigPath returns the path to the user's localconfig.vdf file.
func (u *User) LocalConfigPath() string {
	return filepath.Join(u.ConfigPath, "localconfig.vdf")
}

// HasShortcuts checks if the user has a shortcuts.vdf file.
func (u *User) HasShortcuts() bool {
	_, err := os.Stat(u.ShortcutsPath())
	return err == nil
}

// GetUserByID returns a user by their Steam3 ID.
func (s *Steam) GetUserByID(id string) (*User, error) {
	users, err := s.GetUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.ID == id {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("user not found: %s", id)
}
