package state_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jslay88/zladxhd-installer/internal/state"
)

var _ = Describe("Manager", func() {
	var tmpDir string
	var originalXDG string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "state-test-*")
		Expect(err).NotTo(HaveOccurred())

		originalXDG = os.Getenv("XDG_DATA_HOME")
		_ = os.Setenv("XDG_DATA_HOME", tmpDir)
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
		if originalXDG != "" {
			_ = os.Setenv("XDG_DATA_HOME", originalXDG)
		} else {
			_ = os.Unsetenv("XDG_DATA_HOME")
		}
	})

	Describe("NewManager", func() {
		It("should create manager with correct paths", func() {
			mgr, err := state.NewManager()
			Expect(err).NotTo(HaveOccurred())

			expectedDir := filepath.Join(tmpDir, "zladxhd-installer")
			Expect(mgr.BaseDir()).To(Equal(expectedDir))
		})

		It("should create cache directory", func() {
			_, err := state.NewManager()
			Expect(err).NotTo(HaveOccurred())

			cacheDir := filepath.Join(tmpDir, "zladxhd-installer", "cache")
			info, err := os.Stat(cacheDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
		})

		It("should initialize empty config when none exists", func() {
			mgr, err := state.NewManager()
			Expect(err).NotTo(HaveOccurred())

			Expect(mgr.Config()).NotTo(BeNil())
		})
	})

	Describe("Config", func() {
		It("should save and load config", func() {
			mgr, err := state.NewManager()
			Expect(err).NotTo(HaveOccurred())

			err = mgr.UpdateConfig(func(cfg *state.Config) {
				cfg.LastInstallDir = "/test/install"
				cfg.LastProton = "Proton 8.0"
				cfg.LastSteamUser = "12345"
			})
			Expect(err).NotTo(HaveOccurred())

			// Create new manager to reload config
			mgr2, err := state.NewManager()
			Expect(err).NotTo(HaveOccurred())

			cfg := mgr2.Config()
			Expect(cfg.LastInstallDir).To(Equal("/test/install"))
			Expect(cfg.LastProton).To(Equal("Proton 8.0"))
			Expect(cfg.LastSteamUser).To(Equal("12345"))
		})
	})

	Describe("CachedArchivePath", func() {
		It("should return correct cache path", func() {
			mgr, err := state.NewManager()
			Expect(err).NotTo(HaveOccurred())

			expectedPath := filepath.Join(tmpDir, "zladxhd-installer", "cache", "ZLADXHD.zip")
			Expect(mgr.CachedArchivePath()).To(Equal(expectedPath))
		})
	})

	Describe("CacheDir", func() {
		It("should return correct cache directory path", func() {
			mgr, err := state.NewManager()
			Expect(err).NotTo(HaveOccurred())

			expectedDir := filepath.Join(tmpDir, "zladxhd-installer", "cache")
			Expect(mgr.CacheDir()).To(Equal(expectedDir))
		})
	})
})

var _ = Describe("InstallState", func() {
	var tmpDir string
	var originalXDG string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "state-test-*")
		Expect(err).NotTo(HaveOccurred())

		originalXDG = os.Getenv("XDG_DATA_HOME")
		_ = os.Setenv("XDG_DATA_HOME", tmpDir)
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
		if originalXDG != "" {
			_ = os.Setenv("XDG_DATA_HOME", originalXDG)
		} else {
			_ = os.Unsetenv("XDG_DATA_HOME")
		}
	})

	Describe("NewInstallState", func() {
		It("should create state with current timestamp", func() {
			mgr, err := state.NewManager()
			Expect(err).NotTo(HaveOccurred())

			installState := mgr.NewInstallState()
			Expect(installState).NotTo(BeNil())
			Expect(installState.StartedAt).NotTo(BeZero())
		})

		It("should initialize empty steps", func() {
			mgr, err := state.NewManager()
			Expect(err).NotTo(HaveOccurred())

			installState := mgr.NewInstallState()
			Expect(installState.Steps).To(BeEmpty())
		})
	})

	Describe("Save and Load State", func() {
		It("should persist state across manager instances", func() {
			mgr, err := state.NewManager()
			Expect(err).NotTo(HaveOccurred())

			installState := mgr.NewInstallState()
			installState.InstallDir = "/test/game"
			installState.SteamUserID = "12345"
			installState.AppID = 123456789

			step := installState.AddStep("download")
			step.Start()
			step.Complete()

			err = mgr.SaveState()
			Expect(err).NotTo(HaveOccurred())

			// Create new manager to reload state
			mgr2, err := state.NewManager()
			Expect(err).NotTo(HaveOccurred())

			loaded := mgr2.State()
			Expect(loaded).NotTo(BeNil())
			Expect(loaded.InstallDir).To(Equal("/test/game"))
			Expect(loaded.SteamUserID).To(Equal("12345"))
			Expect(loaded.AppID).To(Equal(uint32(123456789)))
		})
	})

	Describe("ClearState", func() {
		It("should remove state file and clear memory", func() {
			mgr, err := state.NewManager()
			Expect(err).NotTo(HaveOccurred())

			installState := mgr.NewInstallState()
			installState.InstallDir = "/test"
			_ = mgr.SaveState()

			err = mgr.ClearState()
			Expect(err).NotTo(HaveOccurred())
			Expect(mgr.State()).To(BeNil())

			// Verify state not loaded by new manager
			mgr2, err := state.NewManager()
			Expect(err).NotTo(HaveOccurred())
			Expect(mgr2.State()).To(BeNil())
		})
	})

	Describe("AddStep", func() {
		It("should add step with correct name", func() {
			installState := &state.InstallState{
				Steps: make([]state.Step, 0),
			}

			step := installState.AddStep("download")
			Expect(step.Name).To(Equal("download"))
			Expect(step.Status).To(Equal(state.StepPending))
		})

		It("should append to steps slice", func() {
			installState := &state.InstallState{
				Steps: make([]state.Step, 0),
			}

			installState.AddStep("step1")
			installState.AddStep("step2")

			Expect(installState.Steps).To(HaveLen(2))
		})
	})

	Describe("GetStep", func() {
		It("should find step by name", func() {
			installState := &state.InstallState{
				Steps: []state.Step{
					{Name: "step1", Status: state.StepCompleted},
					{Name: "step2", Status: state.StepPending},
				},
			}

			step := installState.GetStep("step1")
			Expect(step).NotTo(BeNil())
			Expect(step.Status).To(Equal(state.StepCompleted))
		})

		It("should return nil for non-existent step", func() {
			installState := &state.InstallState{
				Steps: make([]state.Step, 0),
			}

			step := installState.GetStep("nonexistent")
			Expect(step).To(BeNil())
		})
	})
})

var _ = Describe("Step", func() {
	Describe("Lifecycle", func() {
		var step *state.Step

		BeforeEach(func() {
			step = &state.Step{
				Name:   "test-step",
				Status: state.StepPending,
			}
		})

		It("should start with pending status", func() {
			Expect(step.Status).To(Equal(state.StepPending))
		})

		Describe("Start", func() {
			It("should set status to running", func() {
				step.Start()
				Expect(step.Status).To(Equal(state.StepRunning))
			})

			It("should set start time", func() {
				step.Start()
				Expect(step.StartedAt).NotTo(BeNil())
			})
		})

		Describe("Complete", func() {
			It("should set status to completed", func() {
				step.Start()
				step.Complete()
				Expect(step.Status).To(Equal(state.StepCompleted))
			})

			It("should set completion time", func() {
				step.Complete()
				Expect(step.CompletedAt).NotTo(BeNil())
			})
		})

		Describe("Fail", func() {
			It("should set status to failed", func() {
				step.Fail(nil)
				Expect(step.Status).To(Equal(state.StepFailed))
			})

			It("should set completion time", func() {
				step.Fail(nil)
				Expect(step.CompletedAt).NotTo(BeNil())
			})

			It("should capture error message", func() {
				step.Fail(fmt.Errorf("test error message"))
				Expect(step.Error).To(Equal("test error message"))
			})
		})

		Describe("Skip", func() {
			It("should set status to skipped", func() {
				step.Skip()
				Expect(step.Status).To(Equal(state.StepSkipped))
			})
		})
	})
})
