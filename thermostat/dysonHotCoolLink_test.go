package thermostat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DysonHotCoolLink", func() {
	var myDysonHotCoolLink DysonHotCoolLink
	BeforeEach(func() {
		myDysonHotCoolLink.DysonAPIEmail = "test@test.com"
		myDysonHotCoolLink.DysonAPIPassword = "testpassword"
		myDysonHotCoolLink.IP = "10.0.0.1"
		myDysonHotCoolLink.Port = "9000"
		myDysonHotCoolLink.Serial = "1234"
		myDysonHotCoolLink.DysonAPIEndpoint = "api.cp.dyson.com"
	})
	Describe("connecting to the device", func() {
		Context("when DysonAPIEmail is not specified", func() {
			JustBeforeEach(func() {
				myDysonHotCoolLink.DysonAPIEmail = ""
			})
			It("should return an error", func() {
				err := myDysonHotCoolLink.Connect()
				Expect(err).ShouldNot(BeNil())
			})
		})
		Context("when DysonAPIPassword is not specified", func() {
			JustBeforeEach(func() {
				myDysonHotCoolLink.DysonAPIPassword = ""
			})
			It("should return an error", func() {
				err := myDysonHotCoolLink.Connect()
				Expect(err).ShouldNot(BeNil())
			})
		})
		Context("when Device IP is not specified", func() {
			JustBeforeEach(func() {
				myDysonHotCoolLink.IP = ""
			})
			It("should return an error", func() {
				err := myDysonHotCoolLink.Connect()
				Expect(err).ShouldNot(BeNil())
			})
		})
		Context("when Device port is not specified", func() {
			JustBeforeEach(func() {
				myDysonHotCoolLink.Port = ""
			})
			It("should return an error", func() {
				err := myDysonHotCoolLink.Connect()
				Expect(err).ShouldNot(BeNil())
			})
		})
		Context("when Device serial is not specified", func() {
			JustBeforeEach(func() {
				myDysonHotCoolLink.Serial = ""
			})
			It("should return an error", func() {
				err := myDysonHotCoolLink.Connect()
				Expect(err).ShouldNot(BeNil())
			})
		})
		Describe("adding dyson intermediate credentials", func() {
			Context("when the Dyson Endpoint is not specified", func() {
				JustBeforeEach(func() {
					myDysonHotCoolLink.DysonAPIEndpoint = ""
				})
				It("should be set to the default and succeed", func() {
					err := myDysonHotCoolLink.Connect()
					Expect(err).ShouldNot(BeNil())
				})
			})
			Context("when the Dyson API is not available", func() {
				JustBeforeEach(func() {
					myDysonHotCoolLink.DysonAPIEndpoint = "localhost:test"
				})
				It("should return an error", func() {
					err := myDysonHotCoolLink.Connect()
					Expect(err).ShouldNot(BeNil())
				})
			})
			Context("when the Dyson API returns a bad response body", func() {
				JustBeforeEach(func() {
					s := httptest.NewServer(
						http.HandlerFunc(
							func(w http.ResponseWriter, r *http.Request) {},
						),
					)
					myDysonHotCoolLink.DysonAPIEndpoint = s.URL
				})
				It("should return an error", func() {
					err := myDysonHotCoolLink.Connect()
					Expect(err).ShouldNot(BeNil())
				})
			})
		})
		Describe("adding dyson API info", func() {
			Context("when the Dyson URL is not specified", func() {
				JustBeforeEach(func() {
					myDysonHotCoolLink.DysonAPIEndpoint = "test:test"
				})
				It("should return an error", func() {
					err := myDysonHotCoolLink.addDysonAPIInfo()
					Expect(err).ShouldNot(BeNil())
				})
			})
			Context("when the serial matches and the local credentials are invalid", func() {
				JustBeforeEach(func() {
					s := httptest.NewServer(
						http.HandlerFunc(
							func(w http.ResponseWriter, r *http.Request) {
								testAPIInfo := []DysonAPIInfo{
									DysonAPIInfo{
										Serial: "1234",
									},
								}
								mockResponse, err := json.Marshal(testAPIInfo)
								Expect(err).Should(BeNil())
								w.Write(mockResponse)
							},
						),
					)
					myDysonHotCoolLink.DysonAPIEndpoint = s.URL
				})
				It("should return an error", func() {
					err := myDysonHotCoolLink.addDysonAPIInfo()
					Expect(err).ShouldNot(BeNil())
				})
			})
		})
	})
	Describe("obtaining the temperature", func() {
		It("should return the temperature", func() {
		})
		It("should return no error", func() {
		})
	})
})
