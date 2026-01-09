// Package protontricks provides protontricks installation and management.
package protontricks

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// InstallMethod represents how protontricks is/should be installed.
type InstallMethod string

const (
	InstallNative  InstallMethod = "native"
	InstallFlatpak InstallMethod = "flatpak"
	InstallNone    InstallMethod = "none"
)

// Installation holds information about a protontricks installation.
type Installation struct {
	Method  InstallMethod
	Path    string // Path to the protontricks executable or flatpak app ID
	Version string
}

// Detect checks if protontricks is installed and returns installation info.
func Detect() (*Installation, error) {
	// First check for native installation
	if path, err := exec.LookPath("protontricks"); err == nil {
		version := getVersion(path)
		return &Installation{
			Method:  InstallNative,
			Path:    path,
			Version: version,
		}, nil
	}

	// Check for flatpak installation
	if isFlatpakInstalled() {
		version := getFlatpakVersion()
		return &Installation{
			Method:  InstallFlatpak,
			Path:    "com.github.Matoking.protontricks",
			Version: version,
		}, nil
	}

	return nil, fmt.Errorf("protontricks not found")
}

// getVersion gets the version of a native protontricks installation.
func getVersion(path string) string {
	cmd := exec.Command(path, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// isFlatpakInstalled checks if protontricks is installed via flatpak.
func isFlatpakInstalled() bool {
	cmd := exec.Command("flatpak", "list", "--app", "--columns=application")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "com.github.Matoking.protontricks")
}

// getFlatpakVersion gets the version of the flatpak protontricks installation.
func getFlatpakVersion() string {
	cmd := exec.Command("flatpak", "run", "com.github.Matoking.protontricks", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// Install attempts to install protontricks.
func Install() (*Installation, error) {
	// Try native package managers first
	if err := tryNativeInstall(); err == nil {
		return Detect()
	}

	// Fall back to flatpak
	if err := tryFlatpakInstall(); err == nil {
		return Detect()
	}

	return nil, fmt.Errorf("failed to install protontricks: please install manually")
}

// tryNativeInstall attempts to install protontricks via system package manager.
func tryNativeInstall() error {
	// Detect package manager
	managers := []struct {
		check   string
		install []string
	}{
		{"pacman", []string{"sudo", "pacman", "-S", "--noconfirm", "protontricks"}},
		{"apt", []string{"sudo", "apt", "install", "-y", "protontricks"}},
		{"dnf", []string{"sudo", "dnf", "install", "-y", "protontricks"}},
		{"zypper", []string{"sudo", "zypper", "install", "-y", "protontricks"}},
	}

	for _, pm := range managers {
		if _, err := exec.LookPath(pm.check); err == nil {
			cmd := exec.Command(pm.install[0], pm.install[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
	}

	return fmt.Errorf("no supported package manager found")
}

// tryFlatpakInstall attempts to install protontricks via flatpak.
func tryFlatpakInstall() error {
	if _, err := exec.LookPath("flatpak"); err != nil {
		return fmt.Errorf("flatpak not found")
	}

	cmd := exec.Command("flatpak", "install", "-y", "flathub", "com.github.Matoking.protontricks")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("flatpak install failed: %w", err)
	}

	return nil
}

// SetupFlatpakAliases creates shell aliases for flatpak protontricks.
// Returns the alias commands that should be added to the shell rc file.
func SetupFlatpakAliases() []string {
	return []string{
		`alias protontricks='flatpak run com.github.Matoking.protontricks'`,
		`alias protontricks-launch='flatpak run --command=protontricks-launch com.github.Matoking.protontricks'`,
	}
}

// CheckFlatpakPermissions checks if the flatpak installation has proper permissions.
func CheckFlatpakPermissions() error {
	// Check for filesystem access
	cmd := exec.Command("flatpak", "info", "--show-permissions", "com.github.Matoking.protontricks")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}

	permissions := out.String()
	required := []string{"~/.local/share/Steam", "~/.steam"}

	for _, perm := range required {
		if !strings.Contains(permissions, perm) {
			return fmt.Errorf("flatpak may need additional filesystem permissions for: %s", perm)
		}
	}

	return nil
}
