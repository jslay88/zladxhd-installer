package protontricks_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jslay88/zladxhd-installer/internal/protontricks"
)

var _ = Describe("Installation", func() {
	Describe("InstallMethod", func() {
		It("should have correct constant values", func() {
			Expect(protontricks.InstallNative).To(Equal(protontricks.InstallMethod("native")))
			Expect(protontricks.InstallFlatpak).To(Equal(protontricks.InstallMethod("flatpak")))
			Expect(protontricks.InstallNone).To(Equal(protontricks.InstallMethod("none")))
		})
	})

	Describe("SetupFlatpakAliases", func() {
		It("should return correct aliases", func() {
			aliases := protontricks.SetupFlatpakAliases()
			Expect(aliases).To(HaveLen(2))
			Expect(aliases[0]).To(ContainSubstring("protontricks"))
			Expect(aliases[0]).To(ContainSubstring("flatpak run"))
			Expect(aliases[1]).To(ContainSubstring("protontricks-launch"))
		})
	})

	// Note: Detect() and Install() are hard to test without actual system state
	// They rely on exec.LookPath and flatpak commands
	// In a real scenario, you might want to use interfaces for dependency injection

	Describe("Installation struct", func() {
		It("should have correct fields", func() {
			install := &protontricks.Installation{
				Method:  protontricks.InstallNative,
				Path:    "/usr/bin/protontricks",
				Version: "1.10.5",
			}

			Expect(install.Method).To(Equal(protontricks.InstallNative))
			Expect(install.Path).To(Equal("/usr/bin/protontricks"))
			Expect(install.Version).To(Equal("1.10.5"))
		})

		It("should work with flatpak method", func() {
			install := &protontricks.Installation{
				Method:  protontricks.InstallFlatpak,
				Path:    "com.github.Matoking.protontricks",
				Version: "1.10.5",
			}

			Expect(install.Method).To(Equal(protontricks.InstallFlatpak))
			Expect(install.Path).To(Equal("com.github.Matoking.protontricks"))
		})
	})
})
