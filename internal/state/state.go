// Package state manages application state and settings storage.
// Data is stored in ~/.local/share/zladxhd-installer/
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	appDirName   = "zladxhd-installer"
	configFile   = "config.json"
	stateFile    = "state.json"
	cacheDirName = "cache"
	archiveFile  = "ZLADXHD.zip"
)

// Config stores user preferences and settings.
type Config struct {
	LastInstallDir string `json:"last_install_dir,omitempty"`
	LastProton     string `json:"last_proton,omitempty"`
	LastSteamUser  string `json:"last_steam_user,omitempty"`
}

// InstallState tracks the installation progress for resume/repair.
type InstallState struct {
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	ArchiveChecksum string     `json:"archive_checksum,omitempty"`
	InstallDir      string     `json:"install_dir,omitempty"`
	SteamUserID     string     `json:"steam_user_id,omitempty"`
	AppID           uint32     `json:"app_id,omitempty"`
	Steps           []Step     `json:"steps"`
}

// Step represents a single installation step.
type Step struct {
	Name        string     `json:"name"`
	Status      StepStatus `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       string     `json:"error,omitempty"`
}

// StepStatus represents the status of an installation step.
type StepStatus string

const (
	StepPending   StepStatus = "pending"
	StepRunning   StepStatus = "running"
	StepCompleted StepStatus = "completed"
	StepFailed    StepStatus = "failed"
	StepSkipped   StepStatus = "skipped"
)

// Manager manages application state and settings.
type Manager struct {
	baseDir  string
	cacheDir string
	config   *Config
	state    *InstallState
}

// NewManager creates a new state manager.
func NewManager() (*Manager, error) {
	baseDir, err := getBaseDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get base directory: %w", err)
	}

	cacheDir := filepath.Join(baseDir, cacheDirName)

	// Ensure directories exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	m := &Manager{
		baseDir:  baseDir,
		cacheDir: cacheDir,
	}

	// Load existing config and state
	if err := m.loadConfig(); err != nil {
		// Config doesn't exist yet, use defaults
		m.config = &Config{}
	}

	if err := m.loadState(); err != nil {
		// State doesn't exist yet, use defaults
		m.state = nil
	}

	return m, nil
}

// getBaseDir returns the application data directory.
func getBaseDir() (string, error) {
	// Check XDG_DATA_HOME first
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, appDirName), nil
}

// BaseDir returns the application data directory path.
func (m *Manager) BaseDir() string {
	return m.baseDir
}

// CacheDir returns the cache directory path.
func (m *Manager) CacheDir() string {
	return m.cacheDir
}

// CachedArchivePath returns the path to the cached game archive.
func (m *Manager) CachedArchivePath() string {
	return filepath.Join(m.cacheDir, archiveFile)
}

// Config returns the current configuration.
func (m *Manager) Config() *Config {
	return m.config
}

// State returns the current installation state.
func (m *Manager) State() *InstallState {
	return m.state
}

// SaveConfig saves the configuration to disk.
func (m *Manager) SaveConfig() error {
	return m.saveJSON(filepath.Join(m.baseDir, configFile), m.config)
}

// SaveState saves the installation state to disk.
func (m *Manager) SaveState() error {
	if m.state == nil {
		return nil
	}
	return m.saveJSON(filepath.Join(m.baseDir, stateFile), m.state)
}

// loadConfig loads the configuration from disk.
func (m *Manager) loadConfig() error {
	config := &Config{}
	if err := m.loadJSON(filepath.Join(m.baseDir, configFile), config); err != nil {
		return err
	}
	m.config = config
	return nil
}

// loadState loads the installation state from disk.
func (m *Manager) loadState() error {
	state := &InstallState{}
	if err := m.loadJSON(filepath.Join(m.baseDir, stateFile), state); err != nil {
		return err
	}
	m.state = state
	return nil
}

// saveJSON saves a value to a JSON file.
func (m *Manager) saveJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// loadJSON loads a value from a JSON file.
func (m *Manager) loadJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

// NewInstallState creates a new installation state.
func (m *Manager) NewInstallState() *InstallState {
	m.state = &InstallState{
		StartedAt: time.Now(),
		Steps:     make([]Step, 0),
	}
	return m.state
}

// ClearState removes the installation state.
func (m *Manager) ClearState() error {
	m.state = nil
	path := filepath.Join(m.baseDir, stateFile)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove state file: %w", err)
	}
	return nil
}

// UpdateConfig updates the configuration with the provided function.
func (m *Manager) UpdateConfig(fn func(*Config)) error {
	fn(m.config)
	return m.SaveConfig()
}

// AddStep adds a new step to the installation state.
func (s *InstallState) AddStep(name string) *Step {
	step := Step{
		Name:   name,
		Status: StepPending,
	}
	s.Steps = append(s.Steps, step)
	return &s.Steps[len(s.Steps)-1]
}

// GetStep returns a step by name.
func (s *InstallState) GetStep(name string) *Step {
	for i := range s.Steps {
		if s.Steps[i].Name == name {
			return &s.Steps[i]
		}
	}
	return nil
}

// StartStep marks a step as running.
func (step *Step) Start() {
	now := time.Now()
	step.Status = StepRunning
	step.StartedAt = &now
}

// Complete marks a step as completed.
func (step *Step) Complete() {
	now := time.Now()
	step.Status = StepCompleted
	step.CompletedAt = &now
}

// Fail marks a step as failed with an error.
func (step *Step) Fail(err error) {
	now := time.Now()
	step.Status = StepFailed
	step.CompletedAt = &now
	if err != nil {
		step.Error = err.Error()
	}
}

// Skip marks a step as skipped.
func (step *Step) Skip() {
	step.Status = StepSkipped
}
