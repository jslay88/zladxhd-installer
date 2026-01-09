package proton_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProton(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Proton Suite")
}
