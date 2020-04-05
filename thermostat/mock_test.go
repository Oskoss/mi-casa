package thermostat_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/oskoss/mi-casa/thermostat"
)

var _ = Describe("Mock", func() {

	Describe("getting the current temperature", func() {
		var testMockThermostat MockThermostat
		Context("the test temperature is set to 90 degrees", func() {
			BeforeEach(func() {
				testMockThermostat.Temperature = 90.00
			})
			It("should return no error", func() {
				_, err := testMockThermostat.CurrentTemp()
				Expect(err).ShouldNot(HaveOccurred())
			})
			It("should return 90 degrees as a pointer to float64", func() {
				temp, _ := testMockThermostat.CurrentTemp()
				Expect(*temp).Should(BeAssignableToTypeOf(float64(1.00)))
			})
		})
	})
	Describe("connecting", func() {
		var testMockThermostat MockThermostat
		It("should return no error", func() {
			err := testMockThermostat.Connect()
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

})
