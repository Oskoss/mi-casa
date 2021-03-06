package switcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

//TasmotaT1 implements the switchDevice interface.
// Notice the SwitchNumber which enables us to select
// a specific switch from the T1 since multiple are present
// within the device.
type TasmotaT1 struct {
	SwitchNumber   int
	URI            string
	CurrentStatus  string
	UpdateWindow   time.Duration
	PhysicalDevice TasmotaT1Status
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

//UpdateStatus returns the status of the Tasmota T1 - either "ON" or "OFF"
// Notice for the Tasmota T1 we have 3 switches which can have status'
// therefore we only return which switch is specified within the Tasmota T1 struct
func (t *TasmotaT1) UpdateStatus() (*string, error) {
	checkAgain := true
	layout := "2006.01.02 15:04:05" //Format from Sonoff --> https://github.com/arendst/Sonoff-Tasmota/wiki/JSON-Status-Responses
	lastCheckedTime, err := time.Parse(layout, t.PhysicalDevice.Time)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err,
			"tasmota": t,
		}).Debugf("Failed to parse Tasmota status time")
	} else {
		checkAgainTime := lastCheckedTime.Add(t.UpdateWindow)
		checkAgain = time.Now().UTC().After(checkAgainTime)
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

		var physicalDeviceResp TasmotaT1Status
		err = json.Unmarshal(respBytes, &physicalDeviceResp)
		if err != nil {
			return nil, err
		}
		t.PhysicalDevice = physicalDeviceResp
		switch t.SwitchNumber {
		case 1:
			t.CurrentStatus = t.PhysicalDevice.POWER1
		case 2:
			t.CurrentStatus = t.PhysicalDevice.POWER2
		case 3:
			t.CurrentStatus = t.PhysicalDevice.POWER3
		default:
			errorString := fmt.Sprintf("SwitchNumber %+v specified is not supported -- only 1,2,3 are supported", t.SwitchNumber)
			log.WithFields(log.Fields{
				"switch":       t,
				"switchNumber": t.SwitchNumber,
			}).Error(errorString)
			return nil, fmt.Errorf(errorString)
		}
		return &t.CurrentStatus, nil
	}
	log.WithFields(log.Fields{
		"switch":              t,
		"data last retrieved": t.PhysicalDevice.Time,
	}).Warn("Using cached data from Tasmota")
	return &t.CurrentStatus, nil
}

//TurnOn attempts to turn the switch "ON"
func (t *TasmotaT1) TurnOn() error {
	_, err := t.UpdateStatus()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update status before turning on")
		return err
	}

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(fmt.Sprintf("%s/cm?cmnd=POWER%d%%20ON", t.URI, t.SwitchNumber))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var switchStatusResp map[string]string
	err = json.Unmarshal(respBytes, &switchStatusResp)
	if err != nil {
		return err
	}
	if status, ok := switchStatusResp["POWER"+strconv.Itoa(t.SwitchNumber)]; ok {
		if status != "ON" {
			errorString := fmt.Sprintf("switch %d not reporting as ON after requesting to be ON", t.SwitchNumber)
			log.WithFields(log.Fields{
				"switchStatusResp": switchStatusResp,
				"switchNumber":     t.SwitchNumber,
			}).Error(errorString)
			return fmt.Errorf(errorString)
		}
	} else {
		errorString := fmt.Sprintf("switch %d not reporting back after requesting to be ON", t.SwitchNumber)
		log.WithFields(log.Fields{
			"switchStatusResp": switchStatusResp,
			"switchNumber":     t.SwitchNumber,
		}).Error(errorString)
		return fmt.Errorf(errorString)
	}
	return nil
}

//TurnOff attempts to turn the switch "OFF"
func (t *TasmotaT1) TurnOff() error {

	_, err := t.UpdateStatus()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update status before turning off")
		return err
	}

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(fmt.Sprintf("%s/cm?cmnd=POWER%d%%20OFF", t.URI, t.SwitchNumber))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var switchStatusResp map[string]string
	err = json.Unmarshal(respBytes, &switchStatusResp)
	if err != nil {
		return err
	}
	if status, ok := switchStatusResp["POWER"+strconv.Itoa(t.SwitchNumber)]; ok {
		if status != "OFF" {
			errorString := fmt.Sprintf("switch %d not reporting as OFF after requesting to be OFF", t.SwitchNumber)
			log.WithFields(log.Fields{
				"switchStatusResp": switchStatusResp,
				"switchNumber":     t.SwitchNumber,
			}).Error(errorString)
			return fmt.Errorf(errorString)
		}
	} else {
		errorString := fmt.Sprintf("switch %d not reporting back after requesting to be OFF", t.SwitchNumber)
		log.WithFields(log.Fields{
			"switchStatusResp": switchStatusResp,
			"switchNumber":     t.SwitchNumber,
		}).Error(errorString)
		return fmt.Errorf(errorString)
	}
	return nil
}
