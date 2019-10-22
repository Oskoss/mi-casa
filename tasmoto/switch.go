package tasmoto

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (h *Switch) CurrentStatus() (*Status, error) {

	layout := "2006.01.02 15:04:05" //Format from Sonoff --> https://github.com/arendst/Sonoff-Tasmota/wiki/JSON-Status-Responses
	lastCheckedTime, err := time.Parse(layout, h.AutomationStatus.Time)

	checkAgainTime := time.Now().Add(time.Duration(-30) * time.Second)
	checkAgain := lastCheckedTime.Before(checkAgainTime)
	if err != nil || checkAgain {
		log.WithFields(log.Fields{
			"switch": h,
		}).Debugf("check switch status")
		timeout := time.Duration(5 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		resp, err := client.Get("http://office-closet.oskoss.com/cm?cmnd=state")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var currentStatus Status

		if err := json.Unmarshal(respBytes, &currentStatus); err != nil {
			return nil, err
		}

		return &currentStatus, nil
	}
	fmt.Println("Using Cached Sonoff Data...")
	return &h.AutomationStatus, nil
}

func (h *Switch) CheckManualOverride() error {
	if !h.ManualOverride {
		currentStatus, err := h.CurrentStatus()
		if err != nil {
			return err
		}
		if currentStatus.POWER1 != h.AutomationStatus.POWER1 {
			currentStatus.POWER1 = h.AutomationStatus.POWER1
			currentStatus.POWER2 = h.AutomationStatus.POWER2
			currentStatus.POWER3 = h.AutomationStatus.POWER3
			h.ManualOverride = true
			h.ManualOverrideStart = time.Now()
			return nil
		} else if currentStatus.POWER2 != h.AutomationStatus.POWER2 {
			currentStatus.POWER1 = h.AutomationStatus.POWER1
			currentStatus.POWER2 = h.AutomationStatus.POWER2
			currentStatus.POWER3 = h.AutomationStatus.POWER3
			h.ManualOverride = true
			h.ManualOverrideStart = time.Now()
			return nil
		} else if currentStatus.POWER3 != h.AutomationStatus.POWER3 {
			currentStatus.POWER1 = h.AutomationStatus.POWER1
			currentStatus.POWER2 = h.AutomationStatus.POWER2
			currentStatus.POWER3 = h.AutomationStatus.POWER3
			h.ManualOverride = true
			h.ManualOverrideStart = time.Now()
			return nil
		}
	} else {
		if time.Now().Sub(h.ManualOverrideStart).Minutes() > h.OverrideTimeLength {
			h.ManualOverride = false
		}
	}

	return nil
}

type Switch struct {
	AutomationStatus    Status
	ManualOverride      bool
	ManualOverrideStart time.Time
	OverrideTimeLength  float64
	URI                 string
}

type Status struct {
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
