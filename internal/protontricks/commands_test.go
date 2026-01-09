package protontricks_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jslay88/zladxhd-installer/internal/protontricks"
)

var _ = Describe("Commands", func() {
	Describe("Runner", func() {
		Describe("NewRunner", func() {
			It("should create a runner with the installation", func() {
				install := &protontricks.Installation{
					Method:  protontricks.InstallNative,
					Path:    "/usr/bin/protontricks",
					Version: "1.10.5",
				}

				runner := protontricks.NewRunner(install)
				Expect(runner).NotTo(BeNil())
			})
		})

		// Note: Most Runner methods execute external commands (protontricks, flatpak)
		// These are difficult to unit test without mocking exec.Command
		// In practice, these would be tested via integration tests

		Describe("InstallVerbOptions", func() {
			It("should have Quiet option", func() {
				opts := protontricks.InstallVerbOptions{
					Quiet:          true,
					SuppressOutput: false,
				}
				Expect(opts.Quiet).To(BeTrue())
				Expect(opts.SuppressOutput).To(BeFalse())
			})

			It("should have SuppressOutput option", func() {
				opts := protontricks.InstallVerbOptions{
					Quiet:          false,
					SuppressOutput: true,
				}
				Expect(opts.Quiet).To(BeFalse())
				Expect(opts.SuppressOutput).To(BeTrue())
			})
		})

		Describe("LaunchOptions", func() {
			It("should have SuppressOutput option", func() {
				opts := protontricks.LaunchOptions{
					SuppressOutput: true,
				}
				Expect(opts.SuppressOutput).To(BeTrue())
			})
		})
	})

	// GetPrefixPath tests would require mocking os.UserHomeDir and os.Stat
	// These are tested via integration tests in practice
})
