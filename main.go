package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/oskoss/mi-casa/dyson"
	"github.com/oskoss/mi-casa/tasmoto"
)

func main() {
	defaultDesiredTemp := 72.0
	dysonEmail := os.Getenv("DYSON_API_EMAIL")
	dysonPassword := os.Getenv("DYSON_API_PASSWORD")
	dysonSerial := os.Getenv("DYSON_SERIAL_NUM")
	dysonIP := os.Getenv("DYSON_IP")
	myDyson := dyson.HotCoolLink{
		IP:               dysonIP,
		Port:             "1883",
		Serial:           dysonSerial,
		DysonAPIEmail:    dysonEmail,
		DysonAPIPassword: dysonPassword,
	}
	err := myDyson.MonitorClimate()
	if err != nil {
		panic(err)
	}

	var HVAC tasmoto.Switch
	userSetOverrideTimeString := os.Getenv("MANUAL_OVERRIDE_TIME_MINS")
	if userSetOverrideTimeString == "" {
		userSetOverrideTimeString = "60"
		fmt.Printf("MANUAL_OVERRIDE_TIME_MINS not set in env - Using default value of \"%s\"\n", userSetOverrideTimeString)
	}
	userSetOverrideTime, err := strconv.ParseFloat(userSetOverrideTimeString, 64)
	if err != nil {
		panic(err)
	}
	HVAC.OverrideTimeLength = userSetOverrideTime
	currentStatus, err := HVAC.CurrentStatus()
	if err != nil {
		panic(err)
	}
	HVAC.AutomationStatus = *currentStatus
	HVAC.ManualOverride = false

	errChan := make(chan error)
	tempChan := make(chan float64)
	go ensureTemperature(&HVAC, &myDyson, errChan, tempChan)
	tempChan <- defaultDesiredTemp

	for {
		setTempErr := <-errChan
		fmt.Printf("Error occured when setting temperature: %s", setTempErr.Error())
	}
}

func ensureTemperature(HVAC *tasmoto.Switch, homeSensor *dyson.HotCoolLink, errorChan chan error, tempChan chan float64) {
	tempPadding := 2.0
	desiredTemp := <-tempChan
	for {
		select {
		case newTemp := <-tempChan:
			fmt.Printf("recieved new desired temperature: %f", newTemp)
			desiredTemp = newTemp
		default:
			err := HVAC.CheckManualOverride()
			if err != nil {
				errorChan <- err
			}
			if !HVAC.ManualOverride && homeSensor.ClimateStatus.Data.Tact != "" {
				curTemp, err := strconv.ParseFloat(homeSensor.ClimateStatus.Data.Tact, 64)
				if err != nil {
					errorChan <- err
				}
				curTemp = (curTemp/10-273.15)*9/5 + 32
				fmt.Println("\ncurTemp")
				fmt.Println(curTemp)
				var HVACStatus *tasmoto.Status
				HVACStatus, err = HVAC.CurrentStatus()
				if err != nil {
					errorChan <- err
					HVACStatus.POWER3 = "OFF"
				}
				if (curTemp >= desiredTemp+tempPadding) && (HVACStatus.POWER3 != "ON") {
					timeout := time.Duration(5 * time.Second)
					client := &http.Client{
						Timeout: timeout,
					}
					_, err := client.Get("http://office-closet.oskoss.com/cm?cmnd=Power3%20On")
					if err != nil {
						errorChan <- err
					}
				} else if curTemp <= desiredTemp-tempPadding {
					timeout := time.Duration(5 * time.Second)
					client := &http.Client{
						Timeout: timeout,
					}
					_, err := client.Get("http://office-closet.oskoss.com/cm?cmnd=Power3%20Off")
					if err != nil {
						errorChan <- err
					}
				}

			} else {
				time.Sleep(30 * time.Second)
			}
		}
	}
}
