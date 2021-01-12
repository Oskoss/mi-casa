package switcher_test

import (
	"encoding/json"
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
		})
	})
	Describe("TurnON", func() {
		Context("with a valid OFF switch object", func() {
			var (
				myTasmota TasmotaT1
				err       error
			)
			BeforeEach(func() {
				switchStatusBytes, err := ioutil.ReadFile("../assets/testTasmotaStatus.json")
				Expect(err).Should(BeNil())
				err = json.Unmarshal(switchStatusBytes, &myTasmota.PhysicalDevice)
				Expect(err).Should(BeNil())
				myTasmota.SwitchNumber = 1
				myTasmota.UpdateWindow = time.Duration(5) * time.Second
			})

			Context("with a valid physical switch webserver", func() {
				var server *ghttp.Server
				BeforeEach(func() {
					server = ghttp.NewServer()
					myTasmota.URI = server.URL()

				})
				AfterEach(func() {
					server.Close()
				})
				Context("when in the OFF state", func() {
					var (
						switchOnStatus       map[string]interface{}
						switchOffStatus      map[string]interface{}
						switchStatusBytes    []byte
						switchOnStatusBytes  []byte
						switchOffStatusBytes []byte
					)
					BeforeEach(func() {
						switchStatusBytes, err = ioutil.ReadFile("../assets/testTasmotaStatus.json")
						Expect(err).Should(BeNil())
						err = json.Unmarshal(switchStatusBytes, &switchOnStatus)
						Expect(err).Should(BeNil())
						switchOnStatus["POWER1"] = "ON"
						switchOnStatus["Time"] = time.Now().Format("2006.01.02 15:04:05")
						err = json.Unmarshal(switchStatusBytes, &switchOffStatus)
						Expect(err).Should(BeNil())
						switchOffStatus["POWER1"] = "OFF"
						switchOnStatus["Time"] = time.Now().Format("2006.01.02 15:04:05")

						switchOnStatusBytes, err = json.Marshal(switchOnStatus)
						Expect(err).Should(BeNil())
						switchOffStatusBytes, err = json.Marshal(switchOffStatus)
						Expect(err).Should(BeNil())
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/cm", "cmnd=state"),
								ghttp.RespondWith(http.StatusOK, switchOffStatusBytes),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/cm", "cmnd=POWER1%20ON"),
								ghttp.RespondWith(http.StatusOK, `{"POWER1": "ON"}`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/cm", "cmnd=state"),
								ghttp.RespondWith(http.StatusOK, switchOnStatusBytes),
							),
						)
					})
					It("should turn the switch ON", func() {
						err := myTasmota.TurnOn()
						Expect(err).ShouldNot(HaveOccurred())
						Expect(myTasmota.CurrentStatus).Should(BeEquivalentTo("OFF"))
						Expect(myTasmota.PhysicalDevice.POWER1).Should(BeEquivalentTo("OFF"))
						_, err = myTasmota.UpdateStatus()
						Expect(myTasmota.CurrentStatus).Should(BeEquivalentTo("ON"))
						Expect(myTasmota.PhysicalDevice.POWER3).Should(BeEquivalentTo("ON"))
					})
				})
			})

		})
	})
	Describe("TurnOff", func() {
		Context("with a valid ON switch object", func() {
			var (
				myTasmota TasmotaT1
				err       error
			)
			BeforeEach(func() {
				switchStatusBytes, err := ioutil.ReadFile("../assets/testTasmotaStatus.json")
				Expect(err).Should(BeNil())
				err = json.Unmarshal(switchStatusBytes, &myTasmota.PhysicalDevice)
				Expect(err).Should(BeNil())
				myTasmota.SwitchNumber = 3
				myTasmota.UpdateWindow = time.Duration(5) * time.Second
			})

			Context("with a valid physical switch webserver", func() {
				var server *ghttp.Server
				BeforeEach(func() {
					server = ghttp.NewServer()
					myTasmota.URI = server.URL()

				})
				AfterEach(func() {
					server.Close()
				})
				Context("when in the ON state", func() {
					var (
						switchOnStatus       map[string]interface{}
						switchOffStatus      map[string]interface{}
						switchStatusBytes    []byte
						switchOnStatusBytes  []byte
						switchOffStatusBytes []byte
					)
					BeforeEach(func() {
						switchStatusBytes, err = ioutil.ReadFile("../assets/testTasmotaStatus.json")
						Expect(err).Should(BeNil())
						err = json.Unmarshal(switchStatusBytes, &switchOnStatus)
						Expect(err).Should(BeNil())
						switchOnStatus["POWER3"] = "ON"
						switchOnStatus["Time"] = time.Now().Format("2006.01.02 15:04:05")
						err = json.Unmarshal(switchStatusBytes, &switchOffStatus)
						Expect(err).Should(BeNil())
						switchOffStatus["POWER3"] = "OFF"
						switchOnStatus["Time"] = time.Now().Format("2006.01.02 15:04:05")

						switchOnStatusBytes, err = json.Marshal(switchOnStatus)
						Expect(err).Should(BeNil())
						switchOffStatusBytes, err = json.Marshal(switchOffStatus)
						Expect(err).Should(BeNil())
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/cm", "cmnd=state"),
								ghttp.RespondWith(http.StatusOK, switchOnStatusBytes),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/cm", "cmnd=POWER3%20OFF"),
								ghttp.RespondWith(http.StatusOK, `{"POWER3": "OFF"}`),
							),
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/cm", "cmnd=state"),
								ghttp.RespondWith(http.StatusOK, switchOffStatusBytes),
							),
						)
					})
					It("should turn the switch OFF", func() {
						err := myTasmota.TurnOff()
						Expect(err).ShouldNot(HaveOccurred())
						Expect(myTasmota.CurrentStatus).Should(BeEquivalentTo("ON"))
						Expect(myTasmota.PhysicalDevice.POWER3).Should(BeEquivalentTo("ON"))
						_, err = myTasmota.UpdateStatus()
						Expect(myTasmota.CurrentStatus).Should(BeEquivalentTo("OFF"))
						Expect(myTasmota.PhysicalDevice.POWER3).Should(BeEquivalentTo("OFF"))
					})

				})
			})
		})
	})
})
