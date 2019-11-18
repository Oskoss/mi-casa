package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/oskoss/mi-casa/config"
)

var _ = Describe("Yaml", func() {
	validConfig := YamlConfig{
		FileLocation: "../assets/testConfig.yaml",
	}
	invalidConfig := YamlConfig{
		FileLocation: "doesnotexist.yaml",
	}
	Describe("Getting all fields", func() {
		Context("with a valid yaml file", func() {
			It("should parse the yaml file successfully", func() {
				expected := CasaConfig{
					Name: "myCasa",
				}
				miCasaConfig, err := validConfig.GetAllFields()
				Expect(err).To(BeNil())
				Expect(miCasaConfig).To(BeEquivalentTo(&expected))
			})
		})
		Context("with a invalid yaml file", func() {
			It("should fail", func() {
				_, err := invalidConfig.GetAllFields()
				Expect(err).To(Not(BeNil()))
			})
		})
	})
})
