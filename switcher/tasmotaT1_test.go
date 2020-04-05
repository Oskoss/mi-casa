package switcher_test

import (
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/oskoss/mi-casa/switcher"
)

var _ = Describe("Tasmota", func() {
	Describe("CurrentStatus", func() {
		Context("with a valid switch", func() {
			var (
				server          *ghttp.Server
				myTasmota       TasmotaT1
				updateWindowSec int
			)
			BeforeEach(func() {
				statusJSON, err := ioutil.ReadFile("../assets/testTasmotaStatus.json")
				Expect(err).Should(BeNil())
				server = ghttp.NewServer()
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/cm", "cmnd=state"),
						ghttp.RespondWith(http.StatusOK, statusJSON),
					),
				)
				myTasmota.URI = server.URL()
				myTasmota.SwitchNumber = 1
				updateWindowSec = 5
				myTasmota.UpdateWindow = time.Duration(updateWindowSec) * time.Second
			})
			AfterEach(func() {
				server.Close()
			})
			Context("when switchNumber is 1", func() {
				BeforeEach(func() {
					myTasmota.SwitchNumber = 1
				})
				It("should return the status of switch 1", func() {

					status, err := myTasmota.UpdateStatus()
					Expect(err).Should(BeNil())
					Expect(*status).Should(Equal("OFF"))
				})
			})
			Context("when switchNumber is 2", func() {
				BeforeEach(func() {
					myTasmota.SwitchNumber = 2
				})
				It("should return the status of switch 2", func() {
					status, err := myTasmota.UpdateStatus()
					Expect(err).Should(BeNil())
					Expect(*status).Should(Equal("RANDOM"))
				})
			})
			Context("when switchNumber is 3", func() {
				BeforeEach(func() {
					myTasmota.SwitchNumber = 3
				})
				It("should return the status of switch 3", func() {
					status, err := myTasmota.UpdateStatus()
					Expect(err).ShouldNot(HaveOccurred())
					Expect(*status).Should(Equal("ON"))
				})
			})
			Context("when URI is invalid", func() {
				BeforeEach(func() {
					myTasmota.URI = "invalid"
				})
				It("should return an error", func() {
					status, err := myTasmota.UpdateStatus()
					Expect(status).Should(BeNil())
					Expect(err).Should(HaveOccurred())
				})
			})
			Context("when SwitchNumber is invalid", func() {
				BeforeEach(func() {
					myTasmota.SwitchNumber = 0
				})
				It("should return an error", func() {
					status, err := myTasmota.UpdateStatus()
					Expect(status).Should(BeNil())
					Expect(err).Should(HaveOccurred())
				})
			})
			Context("when the device status was already retrieved within the update window", func() {
				BeforeEach(func() {
					_, err := myTasmota.UpdateStatus()
					Expect(err).ShouldNot(HaveOccurred())
					Expect(server.ReceivedRequests()).Should(HaveLen(1))
					layout := "2006.01.02 15:04:05" //Format from Sonoff --> https://github.com/arendst/Sonoff-Tasmota/wiki/JSON-Status-Responses
					myTasmota.PhysicalDevice.Time = time.Now().UTC().Format(layout)
				})
				It("should return the cached data and not reach out to the device", func() {
					_, err := myTasmota.UpdateStatus()
					Expect(err).ShouldNot(HaveOccurred())
					Expect(server.ReceivedRequests()).Should(HaveLen(1))
				})
			})
			Context("when the device status was retrieved outside the update window", func() {
				BeforeEach(func() {
					myTasmota.PhysicalDevice.Time = ""
					_, err := myTasmota.UpdateStatus()
					Expect(err).ShouldNot(HaveOccurred())
					Expect(server.ReceivedRequests()).Should(HaveLen(1))
					layout := "2006.01.02 15:04:05" //Format from Sonoff --> https://github.com/arendst/Sonoff-Tasmota/wiki/JSON-Status-Responses
					myTasmota.PhysicalDevice.Time = time.Now().UTC().Add(time.Duration(-updateWindowSec) * time.Second).Format(layout)
					statusJSON, err := ioutil.ReadFile("../assets/testTasmotaStatus.json")
					Expect(err).Should(BeNil())
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/cm", "cmnd=state"),
							ghttp.RespondWith(http.StatusOK, statusJSON),
						),
					)
				})
				It("should reach out to the device", func() {
					_, err := myTasmota.UpdateStatus()
					Expect(err).ShouldNot(HaveOccurred())
					Expect(server.ReceivedRequests()).Should(HaveLen(2))
				})
			})
			Context("when physical switch is ON", func() {
				BeforeEach(func() {
					statusJSON, err := ioutil.ReadFile("../assets/testTasmotaStatus.json")
					Expect(err).Should(BeNil())
					server = ghttp.NewServer()
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/cm", "cmnd=state"),
							ghttp.RespondWith(http.StatusOK, statusJSON),
						),
					)
					myTasmota.URI = server.URL()
					myTasmota.SwitchNumber = 3
				})
				Context("but the switch should be OFF according to automation", func() {
					BeforeEach(func() {
						myTasmota.AutomationStatus = "OFF"
					})
					It("should set the switch override status to True", func() {
						_, err := myTasmota.UpdateStatus()
						Expect(err).ShouldNot(HaveOccurred())
						Expect(myTasmota.ManualOverrideStatus).Should(BeTrue())
					})
					It("should start a timer for the override", func() {
						_, err := myTasmota.UpdateStatus()
						Expect(err).ShouldNot(HaveOccurred())
						Expect(myTasmota.ManualOverrideStartTime).Should(BeTemporally("~", time.Now(), time.Duration(5)*time.Second))
					})
				})
			})
			Context("when physical switch is OFF", func() {
				BeforeEach(func() {
					statusJSON, err := ioutil.ReadFile("../assets/testTasmotaStatus.json")
					Expect(err).Should(BeNil())
					server = ghttp.NewServer()
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/cm", "cmnd=state"),
							ghttp.RespondWith(http.StatusOK, statusJSON),
						),
					)
					myTasmota.URI = server.URL()
					myTasmota.SwitchNumber = 1
				})
				Context("but the switch should be ON according to automation", func() {
					BeforeEach(func() {
						myTasmota.AutomationStatus = "ON"
					})
					It("should set the switch override status to True", func() {
						_, err := myTasmota.UpdateStatus()
						Expect(err).ShouldNot(HaveOccurred())
						Expect(myTasmota.ManualOverrideStatus).Should(BeTrue())
					})
					It("should start a timer for the override", func() {
						_, err := myTasmota.UpdateStatus()
						Expect(err).ShouldNot(HaveOccurred())
						Expect(myTasmota.ManualOverrideStartTime).Should(BeTemporally("~", time.Now(), time.Duration(5)*time.Second))
					})
				})
			})
		})
	})
})
