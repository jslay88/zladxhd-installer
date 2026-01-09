package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/jslay88/zladxhd-installer/internal/archive"
	"github.com/jslay88/zladxhd-installer/internal/backup"
	"github.com/jslay88/zladxhd-installer/internal/patcher"
	"github.com/jslay88/zladxhd-installer/internal/proton"
	"github.com/jslay88/zladxhd-installer/internal/protontricks"
	"github.com/jslay88/zladxhd-installer/internal/state"
	"github.com/jslay88/zladxhd-installer/internal/steam"
)

var (
	archivePath string
	installDir  string
	protonName  string
	noBackup    bool
	forceBackup bool
)

var rootCmd = &cobra.Command{
	Use:   "zladxhd-installer",
	Short: "Automated installer for Zelda: Link's Awakening DX HD on Linux with Proton",
	Long: `ZLADXHD Installer automates the manual installation process for
Zelda: Link's Awakening DX HD on Linux using Steam and Proton.

This tool will:
- Install protontricks if needed
- Download/extract the game archive
- Add the game to Steam as a non-Steam game
- Configure Proton compatibility
- Install required .NET runtime
- Download and run the HD patcher`,
	RunE: runInstall,
}

func init() {
	rootCmd.Flags().StringVarP(&archivePath, "archive", "a", "", "Path or URL to game archive (uses cache if not provided)")
	rootCmd.Flags().StringVarP(&installDir, "install-dir", "d", "", "Installation directory (default: ~/.local/share/Steam/steamapps/common/ZLADXHD)")
	rootCmd.Flags().StringVarP(&protonName, "proton", "p", "Proton 10.0", "Proton version to use")
	rootCmd.Flags().BoolVar(&noBackup, "no-backup", false, "Skip Steam backup prompt")
	rootCmd.Flags().BoolVar(&forceBackup, "backup", false, "Force Steam backup without prompt")
}

func Execute() error {
	return rootCmd.Execute()
}

func runInstall(cmd *cobra.Command, args []string) error {
	fmt.Println("🎮 ZLADXHD Installer")
	fmt.Println("==================")
	fmt.Println()

	// Initialize state manager
	stateMgr, err := state.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize state manager: %w", err)
	}

	// Step 1: Check/Install protontricks
	fmt.Println("📦 Checking protontricks...")
	ptInstall, err := ensureProtontricks()
	if err != nil {
		return err
	}
	fmt.Printf("   ✓ protontricks %s (%s)\n", ptInstall.Version, ptInstall.Method)
	fmt.Println()

	// Step 2: Get the game archive
	fmt.Println("📁 Getting game archive...")
	archiveFile, err := getArchive(archivePath, stateMgr)
	if err != nil {
		return err
	}
	fmt.Printf("   ✓ Archive ready: %s\n", archiveFile)
	fmt.Println()

	// Step 3: Discover Steam
	fmt.Println("🔍 Discovering Steam installation...")
	steamInstall, err := steam.Discover()
	if err != nil {
		return fmt.Errorf("failed to find Steam: %w", err)
	}
	fmt.Printf("   ✓ Found Steam at: %s\n", steamInstall.Path)
	fmt.Println()

	// Step 4: Select Steam user
	fmt.Println("👤 Selecting Steam user...")
	user, err := selectSteamUser(steamInstall, stateMgr)
	if err != nil {
		return err
	}
	fmt.Printf("   ✓ Selected user: %s\n", user.DisplayName())
	fmt.Println()

	// Step 5: Backup Steam (optional)
	if err := handleBackup(steamInstall); err != nil {
		return err
	}

	// Step 6: Kill Steam if running
	fmt.Println("🛑 Checking Steam process...")
	if steam.IsRunning() {
		fmt.Println("   Steam is running. Shutting down...")
		if err := steam.Kill(); err != nil {
			return fmt.Errorf("failed to stop Steam: %w", err)
		}
		fmt.Println("   ✓ Steam stopped")
	} else {
		fmt.Println("   ✓ Steam is not running")
	}
	fmt.Println()

	// Step 7: Extract game
	fmt.Println("📦 Extracting game archive...")
	gameDir, err := extractGame(archiveFile, steamInstall, installDir)
	if err != nil {
		return err
	}
	fmt.Printf("   ✓ Extracted to: %s\n", gameDir)
	fmt.Println()

	// Step 8: Find game executable
	exePath, err := findGameExecutable(gameDir)
	if err != nil {
		return err
	}

	// Step 9: Add non-Steam game
	fmt.Println("🎮 Adding game to Steam...")
	shortcut := steam.NewShortcut("Zelda: Link's Awakening DX HD", exePath)
	appID, isNew, err := steam.AddShortcut(user, shortcut)
	if err != nil {
		return fmt.Errorf("failed to add shortcut: %w", err)
	}
	if isNew {
		fmt.Printf("   ✓ Added with AppID: %d\n", appID)
	} else {
		fmt.Printf("   ✓ Already exists with AppID: %d\n", appID)
	}
	fmt.Println()

	// Step 10: Configure Proton
	fmt.Println("⚙️  Configuring Proton...")
	protonCfg, err := configureProton(steamInstall, user, appID, protonName)
	if err != nil {
		return err
	}
	fmt.Printf("   ✓ Using: %s\n", protonCfg.ProtonName)
	fmt.Println()

	// Step 11: Initialize Wine prefix
	fmt.Println("🍷 Initializing Wine prefix...")
	fmt.Println("   This may take a minute on first run...")

	// Show spinner while initializing (suppress Wine debug output)
	prefixSpinner := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription("   Initializing"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSpinnerType(14),
	)
	prefixDone := make(chan error, 1)
	go func() {
		prefixDone <- protonCfg.InitializePrefix(true)
	}()

	// Animate spinner while waiting
	for {
		select {
		case err := <-prefixDone:
			_ = prefixSpinner.Finish()
			fmt.Fprint(os.Stderr, "\r\033[K") // Clear spinner line
			if err != nil {
				return fmt.Errorf("failed to initialize Wine prefix: %w", err)
			}
			fmt.Printf("   ✓ Prefix initialized at: %s\n", protonCfg.PrefixPath())
			goto prefixComplete
		default:
			_ = prefixSpinner.Add(1)
			time.Sleep(100 * time.Millisecond)
		}
	}
prefixComplete:
	fmt.Println()

	// Step 12: Install .NET runtime
	fmt.Println("📦 Installing .NET Desktop Runtime 6...")
	fmt.Println("   This may take a few minutes...")
	ptRunner := protontricks.NewRunner(ptInstall)

	// Show spinner while installing .NET (suppress Wine debug output)
	dotnetSpinner := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription("   Installing"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSpinnerType(14),
	)
	dotnetDone := make(chan error, 1)
	go func() {
		dotnetDone <- ptRunner.InstallDotNetDesktop6(appID, true)
	}()

	// Animate spinner while waiting
	for {
		select {
		case err := <-dotnetDone:
			_ = dotnetSpinner.Finish()
			fmt.Fprint(os.Stderr, "\r\033[K") // Clear spinner line
			if err != nil {
				return fmt.Errorf("failed to install .NET: %w", err)
			}
			fmt.Println("   ✓ .NET Desktop Runtime 6 installed")
			goto dotnetComplete
		default:
			_ = dotnetSpinner.Add(1)
			time.Sleep(100 * time.Millisecond)
		}
	}
dotnetComplete:
	fmt.Println()

	// Step 13: Download patcher
	fmt.Println("⬇️  Downloading HD patcher...")
	p := patcher.NewPatcher(gameDir, stateMgr.CacheDir())
	if err := p.Download(true); err != nil {
		return fmt.Errorf("failed to download patcher: %w", err)
	}
	fmt.Printf("   ✓ Patcher ready: %s\n", filepath.Base(p.PatcherPath))
	fmt.Println()

	// Step 14: Run patcher
	fmt.Println("🔧 Running HD patcher...")
	fmt.Println("   Please follow the patcher instructions in the window that opens.")

	// Show spinner while patcher runs (suppress Wine debug output)
	patcherSpinner := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription("   Running patcher"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSpinnerType(14),
	)
	patcherDone := make(chan error, 1)
	go func() {
		patcherDone <- p.Run(ptRunner, appID, true)
	}()

	// Animate spinner while waiting
	var patcherErr error
	for {
		select {
		case err := <-patcherDone:
			_ = patcherSpinner.Finish()
			fmt.Fprint(os.Stderr, "\r\033[K") // Clear spinner line
			patcherErr = err
			goto patcherComplete
		default:
			_ = patcherSpinner.Add(1)
			time.Sleep(100 * time.Millisecond)
		}
	}
patcherComplete:
	if patcherErr != nil {
		fmt.Printf("   ⚠ Patcher may have exited with error: %v\n", patcherErr)
		fmt.Println("   You can try running it manually later.")
	} else {
		fmt.Println("   ✓ Patcher completed")
	}
	fmt.Println()

	// Save config for next time
	_ = stateMgr.UpdateConfig(func(cfg *state.Config) {
		cfg.LastInstallDir = gameDir
		cfg.LastProton = protonCfg.ProtonName
		cfg.LastSteamUser = user.ID
	})

	// Done!
	fmt.Println("✅ Installation complete!")
	fmt.Println()
	fmt.Println("You can now:")
	fmt.Println("  1. Start Steam")
	fmt.Println("  2. Find 'Zelda: Link's Awakening DX HD' in your library")
	fmt.Println("  3. Play!")
	fmt.Println()

	return nil
}

func ensureProtontricks() (*protontricks.Installation, error) {
	install, err := protontricks.Detect()
	if err == nil {
		return install, nil
	}

	fmt.Println("   protontricks not found. Installing...")
	install, err = protontricks.Install()
	if err != nil {
		return nil, fmt.Errorf("failed to install protontricks: %w", err)
	}

	return install, nil
}

func getArchive(source string, stateMgr *state.Manager) (string, error) {
	cachePath := stateMgr.CachedArchivePath()

	// If no source provided, try to use cache or prompt user
	if source == "" {
		// Check if we have a valid cache
		if archive.FileExists(cachePath) {
			fmt.Println("   Using cached archive...")
			fmt.Println("   Verifying checksum...")
			if err := archive.VerifyExpectedChecksum(cachePath); err == nil {
				return cachePath, nil
			}
			fmt.Println("   ⚠️  Cached archive checksum mismatch, need fresh archive")
		}

		// Prompt user for archive location
		var archiveSource string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Links Awakening DX HD V 1.0.0").
					Description("Enter path or URL to the game archive (search for the title above to find it)").
					Placeholder("/path/to/Links Awakening DX HD.zip or https://...").
					Value(&archiveSource),
			),
		)

		if err := form.Run(); err != nil {
			return "", fmt.Errorf("archive input cancelled: %w", err)
		}

		if archiveSource == "" {
			return "", fmt.Errorf("no archive provided")
		}

		// Recursively call with the provided source
		return getArchive(archiveSource, stateMgr)
	}

	// Check if source is a URL
	if archive.IsURL(source) {
		// Check cache first
		if archive.IsValidCache(cachePath) {
			fmt.Println("   Using cached archive...")
			return cachePath, nil
		}

		// Download to cache
		fmt.Println("   Downloading archive...")
		if err := archive.Download(archive.DownloadOptions{
			URL:          source,
			DestPath:     cachePath,
			ShowProgress: true,
		}); err != nil {
			return "", fmt.Errorf("download failed: %w", err)
		}

		// Verify checksum
		fmt.Println("   Verifying checksum...")
		if err := archive.VerifyExpectedChecksum(cachePath); err != nil {
			_ = os.Remove(cachePath)
			return "", fmt.Errorf("checksum verification failed: %w", err)
		}

		return cachePath, nil
	}

	// Local file
	// Expand ~ in path
	if source[0] == '~' {
		home, _ := os.UserHomeDir()
		source = filepath.Join(home, source[1:])
	}

	// Check if file exists
	if !archive.FileExists(source) {
		return "", fmt.Errorf("archive not found: %s", source)
	}

	// Verify checksum
	fmt.Println("   Verifying checksum...")
	if err := archive.VerifyExpectedChecksum(source); err != nil {
		return "", fmt.Errorf("checksum verification failed: %w", err)
	}

	// Copy to cache if not already there
	if source != cachePath {
		fmt.Println("   Caching archive...")
		if err := archive.CopyFile(source, cachePath); err != nil {
			// Non-fatal, continue with source
			fmt.Printf("   Warning: failed to cache archive: %v\n", err)
			return source, nil
		}
	}

	return cachePath, nil
}

func selectSteamUser(s *steam.Steam, stateMgr *state.Manager) (*steam.User, error) {
	users, err := s.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get Steam users: %w", err)
	}

	if len(users) == 1 {
		return &users[0], nil
	}

	// Check if we have a previously selected user
	lastUserID := stateMgr.Config().LastSteamUser
	if lastUserID != "" {
		for _, u := range users {
			if u.ID == lastUserID {
				var useLastUser bool
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title(fmt.Sprintf("Use previous user: %s?", u.DisplayName())).
							Value(&useLastUser),
					),
				)
				if err := form.Run(); err == nil && useLastUser {
					return &u, nil
				}
				break
			}
		}
	}

	// Build options for selection
	var options []huh.Option[string]
	for _, u := range users {
		options = append(options, huh.NewOption(u.DisplayName(), u.ID))
	}

	var selectedID string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Steam user").
				Options(options...).
				Value(&selectedID),
		),
	)

	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("user selection cancelled: %w", err)
	}

	for _, u := range users {
		if u.ID == selectedID {
			return &u, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

func handleBackup(s *steam.Steam) error {
	if noBackup {
		return nil
	}

	doBackup := forceBackup
	if !forceBackup {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Create a backup of Steam directory before making changes?").
					Description("This will backup ~/.local/share/Steam (excluding steamapps)").
					Value(&doBackup),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("backup prompt cancelled: %w", err)
		}
	}

	if !doBackup {
		fmt.Println("⏭️  Skipping backup")
		fmt.Println()
		return nil
	}

	fmt.Println("💾 Creating Steam backup...")
	fmt.Println("   (excluding steamapps - this may take a minute)")
	opts := backup.DefaultOptions(s.Path)

	// Create progress bar for backup
	bar := progressbar.NewOptions64(
		-1, // Unknown total - will use spinner mode
		progressbar.OptionSetDescription("   Backing up"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(30),
		progressbar.OptionThrottle(100),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
	)

	opts.OnProgress = func(current, total int64, file string) {
		_ = bar.Set64(current)
	}

	result, err := backup.Create(opts)
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	_ = bar.Finish()
	fmt.Printf("   ✓ Backup created: %s (%s, %d files)\n", result.Path, backup.FormatSize(result.Size), result.FileCount)
	fmt.Println()
	return nil
}

func extractGame(archivePath string, s *steam.Steam, customDir string) (string, error) {
	// Determine destination directory
	destDir := customDir
	if destDir == "" {
		destDir = filepath.Join(s.CommonPath(), "ZLADXHD")
	}

	// Expand ~ in path
	if destDir[0] == '~' {
		home, _ := os.UserHomeDir()
		destDir = filepath.Join(home, destDir[1:])
	}

	// Check if already extracted
	if archive.FileExists(destDir) {
		entries, _ := os.ReadDir(destDir)
		if len(entries) > 0 {
			var reExtract bool
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Game directory already exists. Re-extract?").
						Description(destDir).
						Value(&reExtract),
				),
			)

			if err := form.Run(); err != nil || !reExtract {
				return destDir, nil
			}

			// Remove existing directory
			_ = os.RemoveAll(destDir)
		}
	}

	// Extract with StripComponents=1 to remove the root "Links Awakening DX HD" directory
	result, err := archive.Extract(archive.ExtractOptions{
		ArchivePath:     archivePath,
		DestDir:         destDir,
		ShowProgress:    true,
		StripComponents: 1,
	})
	if err != nil {
		return "", fmt.Errorf("extraction failed: %w", err)
	}

	fmt.Printf("   Extracted %d files\n", result.ExtractedFiles)
	return destDir, nil
}

func findGameExecutable(gameDir string) (string, error) {
	// Look for the main game executable
	patterns := []string{
		"Link's Awakening DX HD.exe",
		"LADXHD.exe",
		"*.exe",
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(gameDir, pattern))
		if err != nil {
			continue
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}
			return match, nil
		}
	}

	return "", fmt.Errorf("game executable not found in %s", gameDir)
}

func configureProton(s *steam.Steam, user *steam.User, appID uint32, preferredProton string) (*proton.Config, error) {
	// List available versions
	versions, err := proton.GetAvailableProtonVersions(s)
	if err != nil || len(versions) == 0 {
		return nil, fmt.Errorf("no Proton versions found")
	}

	// Find preferred version index for default selection
	defaultIdx := 0
	preferredVersion, findErr := proton.FindProtonByName(s, preferredProton)
	if findErr == nil {
		for i, v := range versions {
			if v == preferredVersion {
				defaultIdx = i
				break
			}
		}
	}

	// Build options with default selected
	var options []huh.Option[string]
	for _, v := range versions {
		options = append(options, huh.NewOption(v, v))
	}

	// Pre-select the default
	var protonVersion string
	if defaultIdx < len(versions) {
		protonVersion = versions[defaultIdx]
	}

	// Always prompt user for Proton selection
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Proton version").
				Description("Choose which Proton version to use for the game").
				Options(options...).
				Value(&protonVersion),
		),
	)

	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("proton selection cancelled: %w", err)
	}

	cfg, err := proton.NewConfig(s, user, appID, protonVersion)
	if err != nil {
		return nil, err
	}

	// Configure compatibility in config.vdf
	if err := cfg.ConfigureCompatibility(); err != nil {
		fmt.Printf("   Warning: failed to configure compatibility: %v\n", err)
	}

	return cfg, nil
}
