package protontricks

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Runner executes protontricks commands.
type Runner struct {
	install *Installation
}

// NewRunner creates a new protontricks command runner.
func NewRunner(install *Installation) *Runner {
	return &Runner{install: install}
}

// buildCommand builds the appropriate command based on installation method.
func (r *Runner) buildCommand(args ...string) *exec.Cmd {
	if r.install.Method == InstallFlatpak {
		flatpakArgs := append([]string{"run", "com.github.Matoking.protontricks"}, args...)
		return exec.Command("flatpak", flatpakArgs...)
	}
	return exec.Command(r.install.Path, args...)
}

// buildLaunchCommand builds a protontricks-launch command.
func (r *Runner) buildLaunchCommand(args ...string) *exec.Cmd {
	if r.install.Method == InstallFlatpak {
		flatpakArgs := append([]string{"run", "--command=protontricks-launch", "com.github.Matoking.protontricks"}, args...)
		return exec.Command("flatpak", flatpakArgs...)
	}

	// For native install, find protontricks-launch
	launchPath := strings.Replace(r.install.Path, "protontricks", "protontricks-launch", 1)
	return exec.Command(launchPath, args...)
}

// ListGames returns a list of games/apps detected by protontricks.
func (r *Runner) ListGames() (map[string]uint32, error) {
	cmd := r.buildCommand("-l")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list games: %w", err)
	}

	// Parse output: "Game Name (12345)"
	games := make(map[string]uint32)
	lines := strings.Split(out.String(), "\n")
	re := regexp.MustCompile(`^(.+)\s+\((\d+)\)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Found the") {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			name := strings.TrimSpace(matches[1])
			appID, _ := strconv.ParseUint(matches[2], 10, 32)
			games[name] = uint32(appID)
		}
	}

	return games, nil
}

// FindGameByName finds a game's AppID by name substring.
func (r *Runner) FindGameByName(nameSubstr string) (uint32, error) {
	games, err := r.ListGames()
	if err != nil {
		return 0, err
	}

	nameSubstr = strings.ToLower(nameSubstr)
	for name, appID := range games {
		if strings.Contains(strings.ToLower(name), nameSubstr) {
			return appID, nil
		}
	}

	return 0, fmt.Errorf("game not found: %s", nameSubstr)
}

// InstallVerbOptions configures verb installation.
type InstallVerbOptions struct {
	Quiet          bool // Pass -q to winetricks
	SuppressOutput bool // Suppress stdout/stderr (for cleaner CLI output)
}

// InstallVerb installs a winetricks verb into a game's Wine prefix.
func (r *Runner) InstallVerb(appID uint32, verb string, opts InstallVerbOptions) error {
	args := []string{fmt.Sprintf("%d", appID)}
	if opts.Quiet {
		args = append(args, "-q")
	}
	args = append(args, verb)

	cmd := r.buildCommand(args...)

	var outputBuf bytes.Buffer
	if opts.SuppressOutput {
		// Capture output to buffer - dump on error
		cmd.Stdout = &outputBuf
		cmd.Stderr = &outputBuf
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if opts.SuppressOutput && outputBuf.Len() > 0 {
			fmt.Fprintf(os.Stderr, "\n--- Command output (on error) ---\n%s\n--- End output ---\n", outputBuf.String())
		}
		return fmt.Errorf("failed to install %s: %w", verb, err)
	}

	return nil
}

// InstallDotNetDesktop6 installs the .NET Desktop Runtime 6 into a game's Wine prefix.
func (r *Runner) InstallDotNetDesktop6(appID uint32, suppressOutput bool) error {
	return r.InstallVerb(appID, "dotnetdesktop6", InstallVerbOptions{
		Quiet:          true,
		SuppressOutput: suppressOutput,
	})
}

// LaunchOptions configures executable launch.
type LaunchOptions struct {
	SuppressOutput bool     // Suppress stdout/stderr (Wine debug output)
	Args           []string // Additional arguments to pass to the executable
}

// Launch launches an executable in a game's Wine prefix.
func (r *Runner) Launch(appID uint32, exePath string, opts LaunchOptions) error {
	args := []string{
		"--appid", fmt.Sprintf("%d", appID),
		exePath,
	}
	// Append any additional arguments for the executable
	args = append(args, opts.Args...)

	cmd := r.buildLaunchCommand(args...)

	var outputBuf bytes.Buffer
	if opts.SuppressOutput {
		// Capture output to buffer - dump on error
		cmd.Stdout = &outputBuf
		cmd.Stderr = &outputBuf
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.Stdin = os.Stdin

	// Set working directory to the executable's directory
	cmd.Dir = strings.TrimSuffix(exePath, "/"+exePath[strings.LastIndex(exePath, "/")+1:])

	if err := cmd.Run(); err != nil {
		if opts.SuppressOutput && outputBuf.Len() > 0 {
			fmt.Fprintf(os.Stderr, "\n--- Command output (on error) ---\n%s\n--- End output ---\n", outputBuf.String())
		}
		return fmt.Errorf("failed to launch %s: %w", exePath, err)
	}

	return nil
}

// LaunchInDir launches an executable in a specific directory within the Wine prefix.
func (r *Runner) LaunchInDir(appID uint32, exePath string, workDir string, opts LaunchOptions) error {
	args := []string{
		"--appid", fmt.Sprintf("%d", appID),
		exePath,
	}
	// Append any additional arguments for the executable
	args = append(args, opts.Args...)

	cmd := r.buildLaunchCommand(args...)

	var outputBuf bytes.Buffer
	if opts.SuppressOutput {
		// Capture output to buffer - dump on error
		cmd.Stdout = &outputBuf
		cmd.Stderr = &outputBuf
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.Stdin = os.Stdin
	cmd.Dir = workDir

	if err := cmd.Run(); err != nil {
		if opts.SuppressOutput && outputBuf.Len() > 0 {
			fmt.Fprintf(os.Stderr, "\n--- Command output (on error) ---\n%s\n--- End output ---\n", outputBuf.String())
		}
		return fmt.Errorf("failed to launch %s: %w", exePath, err)
	}

	return nil
}

// RunWinetricks runs winetricks directly with custom arguments.
func (r *Runner) RunWinetricks(appID uint32, args ...string) error {
	fullArgs := append([]string{fmt.Sprintf("%d", appID)}, args...)

	cmd := r.buildCommand(fullArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("winetricks command failed: %w", err)
	}

	return nil
}

// GetPrefixPath returns the Wine prefix path for a game.
func (r *Runner) GetPrefixPath(appID uint32) (string, error) {
	// The prefix is typically at ~/.local/share/Steam/steamapps/compatdata/{appid}/pfx
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Check standard location
	prefixPath := fmt.Sprintf("%s/.local/share/Steam/steamapps/compatdata/%d/pfx", home, appID)
	if _, err := os.Stat(prefixPath); err == nil {
		return prefixPath, nil
	}

	// Check flatpak Steam location
	prefixPath = fmt.Sprintf("%s/.var/app/com.valvesoftware.Steam/.local/share/Steam/steamapps/compatdata/%d/pfx", home, appID)
	if _, err := os.Stat(prefixPath); err == nil {
		return prefixPath, nil
	}

	return "", fmt.Errorf("prefix not found for app %d", appID)
}
