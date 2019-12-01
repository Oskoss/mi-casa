package switcher_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSwitcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Switcher Suite")
}
