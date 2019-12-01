package switcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

//TasmotaT1 implements the switchDevice interface.
// Notice the SwitchNumber which enables us to select
// a specific switch from the T1 since multiple are present
// within the device.
type TasmotaT1 struct {
	SwitchNumber    int
	URI             string
	Status          string
	RawDeviceStatus TasmotaT1Status
}

//TasmotaT1Status is the JSON payload received from the device directly
type TasmotaT1Status struct {
	Time      string  `json:"Time"`
	Uptime    string  `json:"Uptime"`
	Vcc       float64 `json:"Vcc"`
	SleepMode string  `json:"SleepMode"`
	Sleep     int     `json:"Sleep"`
	LoadAvg   int     `json:"LoadAvg"`
	POWER1    string  `json:"POWER1"`
	POWER2    string  `json:"POWER2"`
	POWER3    string  `json:"POWER3"`
	Wifi      struct {
		AP        int    `json:"AP"`
		SSID      string `json:"SSId"`
		BSSID     string `json:"BSSId"`
		Channel   int    `json:"Channel"`
		RSSI      int    `json:"RSSI"`
		LinkCount int    `json:"LinkCount"`
		Downtime  string `json:"Downtime"`
	} `json:"Wifi"`
}

//CurrentStatus returns the status of the Tasmota T1 - either "ON" or "OFF"
// Notice for the Tasmota T1 we have 3 switches which can have status'
// therefore we only return which switch is specified within the Tasmota T1 struct
func (t *TasmotaT1) CurrentStatus() (*string, error) {
	checkAgain := true
	layout := "2006.01.02 15:04:05" //Format from Sonoff --> https://github.com/arendst/Sonoff-Tasmota/wiki/JSON-Status-Responses
	lastCheckedTime, err := time.Parse(layout, t.RawDeviceStatus.Time)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err,
			"tasmota": t,
		}).Debugf("Failed to parse Tasmota status time")
	} else {
		checkAgainTime := lastCheckedTime.Add(time.Duration(30) * time.Second)
		checkAgain = checkAgainTime.After(time.Now())
	}
	if checkAgain {
		log.WithFields(log.Fields{
			"switch": t,
		}).Debugf("Checking Tasmota status")
		timeout := time.Duration(5 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		resp, err := client.Get(t.URI + "/cm?cmnd=state")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var currentStatus TasmotaT1Status
		err = json.Unmarshal(respBytes, &currentStatus)
		if err != nil {
			return nil, err
		}
		t.RawDeviceStatus = currentStatus
		switch t.SwitchNumber {
		case 1:
			t.Status = currentStatus.POWER1
		case 2:
			t.Status = currentStatus.POWER2
		case 3:
			t.Status = currentStatus.POWER3
		default:
			errorString := fmt.Sprintf("SwitchNumber %+v specified is not supported -- only 1,2,3 are supported", t.SwitchNumber)
			log.WithFields(log.Fields{
				"switch":       t,
				"switchNumber": t.SwitchNumber,
			}).Error(errorString)
			return nil, fmt.Errorf(errorString)
		}

		return &t.Status, nil
	}
	log.WithFields(log.Fields{
		"switch":              t,
		"data last retrieved": t.RawDeviceStatus.Time,
	}).Warn("Using cached data from Tasmota")
	return &t.Status, nil
}

// func (t *Tasmota) CheckManualOverride() error {
// 	if !h.ManualOverride {
// 		currentStatus, err := h.CurrentStatus()
// 		if err != nil {
// 			return err
// 		}
// 		if currentStatus.POWER1 != h.AutomationStatus.POWER1 {
// 			currentStatus.POWER1 = h.AutomationStatus.POWER1
// 			currentStatus.POWER2 = h.AutomationStatus.POWER2
// 			currentStatus.POWER3 = h.AutomationStatus.POWER3
// 			h.ManualOverride = true
// 			h.ManualOverrideStart = time.Now()
// 			return nil
// 		} else if currentStatus.POWER2 != h.AutomationStatus.POWER2 {
// 			currentStatus.POWER1 = h.AutomationStatus.POWER1
// 			currentStatus.POWER2 = h.AutomationStatus.POWER2
// 			currentStatus.POWER3 = h.AutomationStatus.POWER3
// 			h.ManualOverride = true
// 			h.ManualOverrideStart = time.Now()
// 			return nil
// 		} else if currentStatus.POWER3 != h.AutomationStatus.POWER3 {
// 			currentStatus.POWER1 = h.AutomationStatus.POWER1
// 			currentStatus.POWER2 = h.AutomationStatus.POWER2
// 			currentStatus.POWER3 = h.AutomationStatus.POWER3
// 			h.ManualOverride = true
// 			h.ManualOverrideStart = time.Now()
// 			return nil
// 		}
// 	} else {
// 		if time.Now().Sub(h.ManualOverrideStart).Minutes() > h.OverrideTimeLength {
// 			h.ManualOverride = false
// 		}
// 	}

// 	return nil
// }
