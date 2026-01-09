// Package proton provides Proton configuration and Wine prefix management.
package proton

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jslay88/vdf"
	"github.com/jslay88/zladxhd-installer/internal/steam"
)

// Config holds Proton configuration settings.
type Config struct {
	Steam      *steam.Steam
	User       *steam.User
	AppID      uint32
	ProtonName string
	ProtonPath string
}

// NewConfig creates a new Proton configuration.
func NewConfig(s *steam.Steam, user *steam.User, appID uint32, protonName string) (*Config, error) {
	protonPath, err := s.GetProtonPath(protonName)
	if err != nil {
		return nil, err
	}

	return &Config{
		Steam:      s,
		User:       user,
		AppID:      appID,
		ProtonName: protonName,
		ProtonPath: protonPath,
	}, nil
}

// CompatDataPath returns the path to the Wine prefix for the app.
func (c *Config) CompatDataPath() string {
	return filepath.Join(c.Steam.CompatPath, fmt.Sprintf("%d", c.AppID))
}

// PrefixPath returns the path to the Wine prefix directory.
func (c *Config) PrefixPath() string {
	return filepath.Join(c.CompatDataPath(), "pfx")
}

// CreateCompatData creates the compatdata directory structure for the app.
func (c *Config) CreateCompatData() error {
	compatPath := c.CompatDataPath()

	// Create the compatdata directory
	if err := os.MkdirAll(compatPath, 0755); err != nil {
		return fmt.Errorf("failed to create compatdata directory: %w", err)
	}

	return nil
}

// InitializePrefix initializes the Wine prefix by running Proton.
// This creates the actual Wine prefix structure (drive_c, registry, etc.)
// If suppressOutput is true, Wine debug output is hidden (but dumped on error).
func (c *Config) InitializePrefix(suppressOutput bool) error {
	// Create compatdata directory first
	if err := c.CreateCompatData(); err != nil {
		return err
	}

	// If prefix is already initialized, skip
	if c.HasPrefix() {
		return nil
	}

	// Run Proton with a simple command to initialize the prefix
	// We use "cmd /c exit" which is a Windows command that just exits
	protonExe := filepath.Join(c.ProtonPath, "proton")

	cmd := exec.Command(protonExe, "run", "cmd", "/c", "exit")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("STEAM_COMPAT_CLIENT_INSTALL_PATH=%s", c.Steam.Path),
		fmt.Sprintf("STEAM_COMPAT_DATA_PATH=%s", c.CompatDataPath()),
		"PROTON_LOG=1",
	)

	var outputBuf bytes.Buffer
	if suppressOutput {
		// Capture output to buffer - dump on error
		cmd.Stdout = &outputBuf
		cmd.Stderr = &outputBuf
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		if suppressOutput && outputBuf.Len() > 0 {
			fmt.Fprintf(os.Stderr, "\n--- Proton output (on error) ---\n%s\n--- End output ---\n", outputBuf.String())
		}
		return fmt.Errorf("failed to initialize Wine prefix: %w", err)
	}

	// Verify the prefix was created
	if !c.HasPrefix() {
		if suppressOutput && outputBuf.Len() > 0 {
			fmt.Fprintf(os.Stderr, "\n--- Proton output (prefix not created) ---\n%s\n--- End output ---\n", outputBuf.String())
		}
		return fmt.Errorf("wine prefix was not initialized properly")
	}

	return nil
}

// HasCompatData checks if the compatdata directory exists.
func (c *Config) HasCompatData() bool {
	_, err := os.Stat(c.CompatDataPath())
	return err == nil
}

// HasPrefix checks if the Wine prefix exists and is initialized.
func (c *Config) HasPrefix() bool {
	// Check for system.reg which indicates an initialized prefix
	regPath := filepath.Join(c.PrefixPath(), "system.reg")
	_, err := os.Stat(regPath)
	return err == nil
}

// ConfigureCompatibility sets up Proton compatibility in config.vdf.
// This configures the "Force the use of a specific Steam Play compatibility tool" setting.
func (c *Config) ConfigureCompatibility() error {
	configPath := filepath.Join(c.Steam.Path, "config", "config.vdf")

	// Read or create config.vdf
	var doc *vdf.Document
	var err error

	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		// Create new document
		doc = vdf.NewDocument()
		root := vdf.NewObject("InstallConfigStore")
		doc.AddRoot(root)
	} else {
		doc, err = vdf.ParseFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to parse config.vdf: %w", err)
		}
	}

	// Navigate/create the path: InstallConfigStore -> Software -> Valve -> Steam -> CompatToolMapping
	root := doc.Get("InstallConfigStore")
	if root == nil {
		root = vdf.NewObject("InstallConfigStore")
		doc.AddRoot(root)
	}

	software := root.GetObject("Software")
	if software == nil {
		software = vdf.NewObject("Software")
		root.AddChild(software)
	}

	valve := software.GetObject("Valve")
	if valve == nil {
		valve = vdf.NewObject("Valve")
		software.AddChild(valve)
	}

	steamNode := valve.GetObject("Steam")
	if steamNode == nil {
		steamNode = vdf.NewObject("Steam")
		valve.AddChild(steamNode)
	}

	compatMapping := steamNode.GetObject("CompatToolMapping")
	if compatMapping == nil {
		compatMapping = vdf.NewObject("CompatToolMapping")
		steamNode.AddChild(compatMapping)
	}

	// Create or update app compatibility configuration
	appIDStr := fmt.Sprintf("%d", c.AppID)
	appConfig := compatMapping.GetObject(appIDStr)
	if appConfig == nil {
		appConfig = vdf.NewObject(appIDStr)
		compatMapping.AddChild(appConfig)
	}

	// Get the internal compat tool name
	compatToolName := c.GetCompatToolName()

	// Set the compatibility tool mapping
	appConfig.Set("name", compatToolName)
	appConfig.Set("config", "")
	appConfig.Set("priority", "250")

	// Write the document back
	if err := vdf.WriteFile(configPath, doc); err != nil {
		return fmt.Errorf("failed to write config.vdf: %w", err)
	}

	return nil
}

// GetCompatToolName returns the internal compatibility tool name for Steam's config.
// For official Proton versions, this derives the name from the display name.
// For custom Proton versions (GE-Proton, etc.), the folder name is used directly.
func (c *Config) GetCompatToolName() string {
	// Check if it's in compatibilitytools.d (custom proton)
	customToolPath := filepath.Join(c.Steam.Path, "compatibilitytools.d", c.ProtonName)
	if _, err := os.Stat(customToolPath); err == nil {
		// Custom proton - folder name is the internal name
		return c.ProtonName
	}

	// Official Proton - derive internal name from display name
	// "Proton Experimental" -> "proton_experimental"
	// "Proton 10.0" -> "proton_10"
	// "Proton Hotfix" -> "proton_hotfix"
	name := strings.ToLower(c.ProtonName)
	name = strings.ReplaceAll(name, " - ", " ")
	name = strings.ReplaceAll(name, " ", "_")
	// Remove .0 suffix from version numbers (e.g., "proton_10.0" -> "proton_10")
	if strings.Contains(name, "_") {
		parts := strings.Split(name, "_")
		for i, part := range parts {
			if strings.HasSuffix(part, ".0") {
				parts[i] = strings.TrimSuffix(part, ".0")
			}
		}
		name = strings.Join(parts, "_")
	}
	return name
}

// GetAvailableProtonVersions returns a list of Proton versions, with recommended ones first.
func GetAvailableProtonVersions(s *steam.Steam) ([]string, error) {
	versions, err := s.GetProtonVersions()
	if err != nil {
		return nil, err
	}

	// Sort: Proton Experimental first, then GE-Proton, then others
	var experimental []string
	var ge []string
	var regular []string
	var other []string

	for _, v := range versions {
		switch {
		case strings.Contains(v, "Experimental"):
			experimental = append(experimental, v)
		case strings.HasPrefix(v, "GE-Proton"):
			ge = append(ge, v)
		case strings.HasPrefix(v, "Proton"):
			regular = append(regular, v)
		default:
			other = append(other, v)
		}
	}

	var sorted []string
	sorted = append(sorted, experimental...)
	sorted = append(sorted, ge...)
	sorted = append(sorted, regular...)
	sorted = append(sorted, other...)

	return sorted, nil
}

// FindProtonByName finds a Proton version by partial name match.
func FindProtonByName(s *steam.Steam, name string) (string, error) {
	versions, err := s.GetProtonVersions()
	if err != nil {
		return "", err
	}

	name = strings.ToLower(name)
	nameNorm := normalizeProtonName(name)

	// First try exact match
	for _, v := range versions {
		if strings.ToLower(v) == name {
			return v, nil
		}
	}

	// Try normalized exact match (handles "Proton Experimental" vs "Proton - Experimental")
	for _, v := range versions {
		if normalizeProtonName(strings.ToLower(v)) == nameNorm {
			return v, nil
		}
	}

	// Then try partial match
	for _, v := range versions {
		if strings.Contains(strings.ToLower(v), name) {
			return v, nil
		}
	}

	// Try normalized partial match
	for _, v := range versions {
		if strings.Contains(normalizeProtonName(strings.ToLower(v)), nameNorm) {
			return v, nil
		}
	}

	return "", fmt.Errorf("proton version not found: %s", name)
}

// normalizeProtonName removes dashes and extra spaces for fuzzy matching.
func normalizeProtonName(name string) string {
	// Replace " - " with " "
	name = strings.ReplaceAll(name, " - ", " ")
	// Replace multiple spaces with single space
	for strings.Contains(name, "  ") {
		name = strings.ReplaceAll(name, "  ", " ")
	}
	return strings.TrimSpace(name)
}
