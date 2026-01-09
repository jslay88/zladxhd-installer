package steam

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/jslay88/vdf"
)

// Shortcut represents a non-Steam game shortcut.
type Shortcut struct {
	AppID               uint32
	AppName             string
	Exe                 string
	StartDir            string
	Icon                string
	ShortcutPath        string
	LaunchOptions       string
	IsHidden            uint32
	AllowDesktopConfig  uint32
	AllowOverlay        uint32
	OpenVR              uint32
	Devkit              uint32
	DevkitGameID        string
	DevkitOverrideAppID uint32
	LastPlayTime        uint32
	FlatpakAppID        string
	Tags                map[string]string
}

// NewShortcut creates a new shortcut with default values.
func NewShortcut(appName, exePath string) *Shortcut {
	startDir := filepath.Dir(exePath)
	return &Shortcut{
		AppID:               0, // Will be generated
		AppName:             appName,
		Exe:                 fmt.Sprintf("\"%s\"", exePath),
		StartDir:            fmt.Sprintf("\"%s\"", startDir),
		Icon:                "",
		ShortcutPath:        "",
		LaunchOptions:       "",
		IsHidden:            0,
		AllowDesktopConfig:  1,
		AllowOverlay:        1,
		OpenVR:              0,
		Devkit:              0,
		DevkitGameID:        "",
		DevkitOverrideAppID: 0,
		LastPlayTime:        0,
		FlatpakAppID:        "",
		Tags:                make(map[string]string),
	}
}

// GenerateAppID generates a random AppID in the non-Steam game range.
// Steam uses the range 0xFF000000 to 0xFFFFFFFF for non-Steam games.
func GenerateAppID() (uint32, error) {
	// Range: 0xFF000000 (4278190080) to 0xFFFFFFFF (4294967295)
	minID := uint32(0xFF000000)
	maxID := uint32(0xFFFFFFFF)
	rangeSize := int64(maxID - minID)

	n, err := rand.Int(rand.Reader, big.NewInt(rangeSize))
	if err != nil {
		return 0, fmt.Errorf("failed to generate random AppID: %w", err)
	}

	return minID + uint32(n.Int64()), nil
}

// ReadShortcuts reads shortcuts from a shortcuts.vdf file.
func ReadShortcuts(path string) ([]Shortcut, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Shortcut{}, nil
		}
		return nil, fmt.Errorf("failed to read shortcuts file: %w", err)
	}

	if len(data) == 0 {
		return []Shortcut{}, nil
	}

	vdfMap, err := vdf.ReadBinary(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse shortcuts VDF: %w", err)
	}

	shortcutsObj, ok := vdfMap["shortcuts"]
	if !ok {
		return []Shortcut{}, nil
	}

	shortcutsMap, ok := shortcutsObj.(vdf.Map)
	if !ok {
		return []Shortcut{}, nil
	}

	var shortcuts []Shortcut
	for _, v := range shortcutsMap {
		entryMap, ok := v.(vdf.Map)
		if !ok {
			continue
		}

		shortcut := parseShortcutEntry(entryMap)
		shortcuts = append(shortcuts, shortcut)
	}

	return shortcuts, nil
}

// parseShortcutEntry parses a single shortcut entry from a VDF map.
func parseShortcutEntry(m vdf.Map) Shortcut {
	s := Shortcut{
		Tags: make(map[string]string),
	}

	if v, ok := m["appid"].(uint32); ok {
		s.AppID = v
	}
	if v, ok := m["AppName"].(string); ok {
		s.AppName = v
	}
	if v, ok := m["Exe"].(string); ok {
		s.Exe = v
	}
	if v, ok := m["StartDir"].(string); ok {
		s.StartDir = v
	}
	if v, ok := m["icon"].(string); ok {
		s.Icon = v
	}
	if v, ok := m["ShortcutPath"].(string); ok {
		s.ShortcutPath = v
	}
	if v, ok := m["LaunchOptions"].(string); ok {
		s.LaunchOptions = v
	}
	if v, ok := m["IsHidden"].(uint32); ok {
		s.IsHidden = v
	}
	if v, ok := m["AllowDesktopConfig"].(uint32); ok {
		s.AllowDesktopConfig = v
	}
	if v, ok := m["AllowOverlay"].(uint32); ok {
		s.AllowOverlay = v
	}
	if v, ok := m["OpenVR"].(uint32); ok {
		s.OpenVR = v
	}
	if v, ok := m["Devkit"].(uint32); ok {
		s.Devkit = v
	}
	if v, ok := m["DevkitGameID"].(string); ok {
		s.DevkitGameID = v
	}
	if v, ok := m["DevkitOverrideAppID"].(uint32); ok {
		s.DevkitOverrideAppID = v
	}
	if v, ok := m["LastPlayTime"].(uint32); ok {
		s.LastPlayTime = v
	}
	if v, ok := m["FlatpakAppID"].(string); ok {
		s.FlatpakAppID = v
	}
	if tags, ok := m["tags"].(vdf.Map); ok {
		for k, v := range tags {
			if str, ok := v.(string); ok {
				s.Tags[k] = str
			}
		}
	}

	return s
}

// toVDFMap converts a Shortcut to a VDF map.
func (s *Shortcut) toVDFMap() vdf.Map {
	m := vdf.Map{
		"appid":               s.AppID,
		"AppName":             s.AppName,
		"Exe":                 s.Exe,
		"StartDir":            s.StartDir,
		"icon":                s.Icon,
		"ShortcutPath":        s.ShortcutPath,
		"LaunchOptions":       s.LaunchOptions,
		"IsHidden":            s.IsHidden,
		"AllowDesktopConfig":  s.AllowDesktopConfig,
		"AllowOverlay":        s.AllowOverlay,
		"OpenVR":              s.OpenVR,
		"Devkit":              s.Devkit,
		"DevkitGameID":        s.DevkitGameID,
		"DevkitOverrideAppID": s.DevkitOverrideAppID,
		"LastPlayTime":        s.LastPlayTime,
		"FlatpakAppID":        s.FlatpakAppID,
	}

	tags := vdf.Map{}
	for k, v := range s.Tags {
		tags[k] = v
	}
	m["tags"] = tags

	return m
}

// WriteShortcuts writes shortcuts to a shortcuts.vdf file.
func WriteShortcuts(path string, shortcuts []Shortcut) error {
	shortcutsMap := vdf.Map{}
	for i, s := range shortcuts {
		shortcutsMap[fmt.Sprintf("%d", i)] = s.toVDFMap()
	}

	vdfMap := vdf.Map{
		"shortcuts": shortcutsMap,
	}

	data, err := vdf.WriteBinary(vdfMap)
	if err != nil {
		return fmt.Errorf("failed to write shortcuts VDF: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write shortcuts file: %w", err)
	}

	return nil
}

// AddShortcut adds a shortcut to a user's shortcuts.vdf.
// If a shortcut with the same name already exists, it returns the existing AppID.
// Returns the AppID and a boolean indicating if it was newly created.
func AddShortcut(user *User, shortcut *Shortcut) (uint32, bool, error) {
	shortcuts, err := ReadShortcuts(user.ShortcutsPath())
	if err != nil {
		return 0, false, err
	}

	// Check if shortcut with same name already exists
	for _, s := range shortcuts {
		if s.AppName == shortcut.AppName {
			// Already exists, return existing AppID
			return s.AppID, false, nil
		}
	}

	// Generate AppID if not set
	if shortcut.AppID == 0 {
		appID, err := GenerateAppID()
		if err != nil {
			return 0, false, err
		}

		// Make sure it doesn't conflict with existing shortcuts
		for {
			conflict := false
			for _, s := range shortcuts {
				if s.AppID == appID {
					conflict = true
					break
				}
			}
			if !conflict {
				break
			}
			appID, err = GenerateAppID()
			if err != nil {
				return 0, false, err
			}
		}

		shortcut.AppID = appID
	}

	shortcuts = append(shortcuts, *shortcut)

	if err := WriteShortcuts(user.ShortcutsPath(), shortcuts); err != nil {
		return 0, false, err
	}

	return shortcut.AppID, true, nil
}

// UpdateShortcut updates an existing shortcut by AppID.
// If the shortcut doesn't exist, it returns an error.
func UpdateShortcut(user *User, shortcut *Shortcut) error {
	shortcuts, err := ReadShortcuts(user.ShortcutsPath())
	if err != nil {
		return err
	}

	found := false
	for i, s := range shortcuts {
		if s.AppID == shortcut.AppID {
			shortcuts[i] = *shortcut
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("shortcut with AppID %d not found", shortcut.AppID)
	}

	return WriteShortcuts(user.ShortcutsPath(), shortcuts)
}

// FindShortcutByName finds a shortcut by app name.
func FindShortcutByName(user *User, appName string) (*Shortcut, error) {
	shortcuts, err := ReadShortcuts(user.ShortcutsPath())
	if err != nil {
		return nil, err
	}

	for _, s := range shortcuts {
		if s.AppName == appName {
			return &s, nil
		}
	}

	return nil, nil
}

// RemoveShortcut removes a shortcut by AppID.
func RemoveShortcut(user *User, appID uint32) error {
	shortcuts, err := ReadShortcuts(user.ShortcutsPath())
	if err != nil {
		return err
	}

	var filtered []Shortcut
	for _, s := range shortcuts {
		if s.AppID != appID {
			filtered = append(filtered, s)
		}
	}

	return WriteShortcuts(user.ShortcutsPath(), filtered)
}
