package protontricks_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProtontricks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Protontricks Suite")
}
