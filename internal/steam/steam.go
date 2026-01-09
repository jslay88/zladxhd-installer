// Package steam provides Steam-related operations including path discovery,
// process management, and configuration file handling.
package steam

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// DefaultSteamPath is the default Steam installation path on Linux.
const DefaultSteamPath = ".local/share/Steam"

// Steam represents a Steam installation.
type Steam struct {
	Path         string
	UserDataPath string
	ConfigPath   string
	AppsPath     string
	CompatPath   string
}

// Discover finds the Steam installation on the system.
func Discover() (*Steam, error) {
	steamPath, err := findSteamPath()
	if err != nil {
		return nil, err
	}

	s := &Steam{
		Path:         steamPath,
		UserDataPath: filepath.Join(steamPath, "userdata"),
		ConfigPath:   filepath.Join(steamPath, "config"),
		AppsPath:     filepath.Join(steamPath, "steamapps"),
		CompatPath:   filepath.Join(steamPath, "steamapps", "compatdata"),
	}

	// Verify the installation
	if err := s.verify(); err != nil {
		return nil, err
	}

	return s, nil
}

// findSteamPath looks for the Steam installation directory.
func findSteamPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Check common Steam paths
	paths := []string{
		filepath.Join(home, DefaultSteamPath),
		filepath.Join(home, ".steam", "steam"),
		filepath.Join(home, ".steam", "debian-installation"),
	}

	// Check STEAM_DIR environment variable
	if steamDir := os.Getenv("STEAM_DIR"); steamDir != "" {
		paths = append([]string{steamDir}, paths...)
	}

	// Also check XDG_DATA_HOME
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(home, ".local", "share")
	}
	paths = append(paths, filepath.Join(dataHome, "Steam"))

	for _, path := range paths {
		// Resolve symlinks
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			continue
		}

		if info, err := os.Stat(resolved); err == nil && info.IsDir() {
			// Check for steam.sh or steamapps directory
			if _, err := os.Stat(filepath.Join(resolved, "steamapps")); err == nil {
				return resolved, nil
			}
			if _, err := os.Stat(filepath.Join(resolved, "steam.sh")); err == nil {
				return resolved, nil
			}
		}
	}

	return "", fmt.Errorf("Steam installation not found. Checked paths: %v", paths)
}

// verify checks that the Steam installation is valid.
func (s *Steam) verify() error {
	dirs := []string{s.UserDataPath, s.ConfigPath, s.AppsPath}
	for _, dir := range dirs {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			return fmt.Errorf("Steam directory not found: %s", dir)
		}
	}
	return nil
}

// CommonPath returns the path to Steam's common games directory.
func (s *Steam) CommonPath() string {
	return filepath.Join(s.AppsPath, "common")
}

// GetProtonVersions returns a list of installed Proton versions.
func (s *Steam) GetProtonVersions() ([]string, error) {
	commonPath := s.CommonPath()
	entries, err := os.ReadDir(commonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read common directory: %w", err)
	}

	var versions []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Check for Proton directories
		if strings.HasPrefix(name, "Proton") || strings.HasPrefix(name, "GE-Proton") {
			versions = append(versions, name)
		}
	}

	return versions, nil
}

// GetProtonPath returns the path to a specific Proton version.
func (s *Steam) GetProtonPath(version string) (string, error) {
	protonPath := filepath.Join(s.CommonPath(), version)
	if info, err := os.Stat(protonPath); err != nil || !info.IsDir() {
		return "", fmt.Errorf("proton version not found: %s", version)
	}
	return protonPath, nil
}

// IsRunning checks if Steam is currently running.
func IsRunning() bool {
	cmd := exec.Command("pgrep", "-x", "steam")
	err := cmd.Run()
	return err == nil
}

// GetPID returns the PID of the running Steam process.
func GetPID() (int, error) {
	cmd := exec.Command("pgrep", "-x", "steam")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("Steam is not running")
	}

	pidStr := strings.TrimSpace(string(output))
	// pgrep might return multiple PIDs, take the first one
	if idx := strings.Index(pidStr, "\n"); idx != -1 {
		pidStr = pidStr[:idx]
	}

	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse PID: %w", err)
	}

	return pid, nil
}

// Kill terminates the Steam process gracefully, with fallback to SIGKILL.
func Kill() error {
	pid, err := GetPID()
	if err != nil {
		// Steam not running, nothing to do
		return nil
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// Try SIGTERM first
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	// Wait for process to exit (up to 10 seconds)
	for i := 0; i < 20; i++ {
		time.Sleep(500 * time.Millisecond)
		if !IsRunning() {
			return nil
		}
	}

	// If still running, use SIGKILL
	if err := process.Signal(syscall.SIGKILL); err != nil {
		return fmt.Errorf("failed to send SIGKILL: %w", err)
	}

	// Wait a bit more for the process to die
	time.Sleep(1 * time.Second)
	if IsRunning() {
		return fmt.Errorf("failed to kill Steam process")
	}

	return nil
}

// WaitForExit waits for Steam to exit, with a timeout.
func WaitForExit(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !IsRunning() {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for Steam to exit")
}
