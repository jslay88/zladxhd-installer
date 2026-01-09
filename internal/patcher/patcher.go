// Package patcher handles downloading and running the LADXHD patcher.
package patcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jslay88/zladxhd-installer/internal/archive"
	"github.com/jslay88/zladxhd-installer/internal/protontricks"
)

const (
	// GitHubRepo is the repository for LADXHD patcher releases.
	// See: https://github.com/BigheadSMZ/Zelda-LA-DX-HD-Updated/releases
	GitHubRepo = "BigheadSMZ/Zelda-LA-DX-HD-Updated"
	// GitHubAPIURL is the GitHub API URL for releases.
	GitHubAPIURL = "https://api.github.com/repos/%s/releases/latest"
	// PatcherNamePattern is the pattern to match patcher files.
	PatcherNamePattern = "LADXHD.Patcher"
)

// Release represents a GitHub release.
type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a GitHub release asset.
type Asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int64  `json:"size"`
}

// Patcher manages the LADXHD patcher.
type Patcher struct {
	GameDir     string
	PatcherPath string
	CacheDir    string
}

// NewPatcher creates a new patcher instance.
func NewPatcher(gameDir string, cacheDir string) *Patcher {
	return &Patcher{
		GameDir:  gameDir,
		CacheDir: cacheDir,
	}
}

// GetLatestRelease fetches the latest patcher release info from GitHub.
func GetLatestRelease() (*Release, error) {
	url := fmt.Sprintf(GitHubAPIURL, GitHubRepo)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	return &release, nil
}

// FindPatcherAsset finds the patcher executable in the release assets.
func FindPatcherAsset(release *Release) (*Asset, error) {
	// First, try to find an asset with "patcher" in the name
	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, strings.ToLower(PatcherNamePattern)) && strings.HasSuffix(name, ".exe") {
			return &asset, nil
		}
	}

	// Fallback: find any .exe file
	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if strings.HasSuffix(name, ".exe") {
			return &asset, nil
		}
	}

	// List available assets for debugging
	var assetNames []string
	for _, asset := range release.Assets {
		assetNames = append(assetNames, asset.Name)
	}

	return nil, fmt.Errorf("patcher executable not found in release assets. Available: %v", assetNames)
}

// Download downloads the patcher to the game directory.
func (p *Patcher) Download(showProgress bool) error {
	release, err := GetLatestRelease()
	if err != nil {
		return err
	}

	asset, err := FindPatcherAsset(release)
	if err != nil {
		return err
	}

	// Download to game directory
	p.PatcherPath = filepath.Join(p.GameDir, asset.Name)

	// Check if already downloaded
	if _, err := os.Stat(p.PatcherPath); err == nil {
		// Already exists, verify it's not empty
		info, _ := os.Stat(p.PatcherPath)
		if info.Size() > 0 {
			return nil
		}
	}

	// Download the patcher
	err = archive.Download(archive.DownloadOptions{
		URL:          asset.DownloadURL,
		DestPath:     p.PatcherPath,
		ShowProgress: showProgress,
	})
	if err != nil {
		return fmt.Errorf("failed to download patcher: %w", err)
	}

	return nil
}

// DownloadFromURL downloads the patcher from a specific URL.
func (p *Patcher) DownloadFromURL(url string, showProgress bool) error {
	filename := filepath.Base(url)
	p.PatcherPath = filepath.Join(p.GameDir, filename)

	err := archive.Download(archive.DownloadOptions{
		URL:          url,
		DestPath:     p.PatcherPath,
		ShowProgress: showProgress,
	})
	if err != nil {
		return fmt.Errorf("failed to download patcher: %w", err)
	}

	return nil
}

// Run runs the patcher using protontricks.
func (p *Patcher) Run(runner *protontricks.Runner, appID uint32, suppressOutput bool) error {
	if p.PatcherPath == "" {
		return fmt.Errorf("patcher not downloaded")
	}

	// Run the patcher in the game directory
	return runner.LaunchInDir(appID, p.PatcherPath, p.GameDir, protontricks.LaunchOptions{
		SuppressOutput: suppressOutput,
	})
}

// FindExisting looks for an existing patcher in the game directory.
func (p *Patcher) FindExisting() (string, error) {
	entries, err := os.ReadDir(p.GameDir)
	if err != nil {
		return "", fmt.Errorf("failed to read game directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if strings.Contains(name, strings.ToLower(PatcherNamePattern)) && strings.HasSuffix(name, ".exe") {
			path := filepath.Join(p.GameDir, entry.Name())
			p.PatcherPath = path
			return path, nil
		}
	}

	return "", fmt.Errorf("patcher not found in game directory")
}

// GetDownloadURL constructs the download URL for a specific patcher version.
func GetDownloadURL(version string, filename string) string {
	return fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", GitHubRepo, version, filename)
}

// ListAvailableVersions lists available patcher versions from GitHub.
func ListAvailableVersions() ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", GitHubRepo)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %w", err)
	}

	var versions []string
	for _, r := range releases {
		versions = append(versions, r.TagName)
	}

	return versions, nil
}
