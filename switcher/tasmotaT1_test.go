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
	var (
		server    *ghttp.Server
		myTasmota TasmotaT1
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
	})
	AfterEach(func() {
		server.Close()
	})
	Describe("getting current switch status", func() {
		Context("when switchNumber is 1", func() {
			JustBeforeEach(func() {
				myTasmota.SwitchNumber = 1
			})
			It("should return the status of switch 1", func() {

				status, err := myTasmota.CurrentStatus()
				Expect(err).Should(BeNil())
				Expect(*status).Should(Equal("OFF"))
			})
		})
		Context("when switchNumber is 2", func() {
			JustBeforeEach(func() {
				myTasmota.SwitchNumber = 2
			})
			It("should return the status of switch 2", func() {
				status, err := myTasmota.CurrentStatus()
				Expect(err).Should(BeNil())
				Expect(*status).Should(Equal("RANDOM"))
			})
		})
		Context("when switchNumber is 3", func() {
			JustBeforeEach(func() {
				myTasmota.SwitchNumber = 3
			})
			It("should return the status of switch 3", func() {
				status, err := myTasmota.CurrentStatus()
				Expect(err).Should(BeNil())
				Expect(*status).Should(Equal("ON"))
			})
		})
		Context("when URI is invalid", func() {
			JustBeforeEach(func() {
				myTasmota.URI = "invalid"
			})
			It("should return an error", func() {
				status, err := myTasmota.CurrentStatus()
				Expect(status).Should(BeNil())
				Expect(err).Should(HaveOccurred())
			})
		})
		Context("when SwitchNumber is invalid", func() {
			JustBeforeEach(func() {
				myTasmota.SwitchNumber = 0
			})
			It("should return an error", func() {
				status, err := myTasmota.CurrentStatus()
				Expect(status).Should(BeNil())
				Expect(err).Should(HaveOccurred())
			})
		})
		Context("when the device status was already retrieved less than 30 seconds ago", func() {
			JustBeforeEach(func() {
				//currentStatus.Time must be set to within the last 30 seconds. We use time.Now()
				_, err := myTasmota.CurrentStatus()
				Expect(err).ShouldNot(HaveOccurred())
				layout := "2006.01.02 15:04:05" //Format from Sonoff --> https://github.com/arendst/Sonoff-Tasmota/wiki/JSON-Status-Responses
				myTasmota.RawDeviceStatus.Time = time.Now().Format(layout)
			})
			It("should return the cached data and not reach out to the device", func() {
				_, err := myTasmota.CurrentStatus()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(server.ReceivedRequests()).Should(HaveLen(1))

			})
		})
	})
})
