package thermostat_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestThermostat(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Thermostat Suite")
}
