package home

import (
	"net/http"
	"strconv"
	"time"

	"github.com/oskoss/mi-casa/dyson"
	"github.com/oskoss/mi-casa/tasmoto"
	log "github.com/sirupsen/logrus"
)

func (myHome *Home) SetTemperature(temperature float64) error {
	log.WithFields(log.Fields{
		"temperature": temperature,
	}).Printf("submitting change of temperature")
	myHome.TemperatureChannel <- temperature
	return nil
}

func (myHome *Home) AddHotCoolLink(dysonIP, dysonSerial, dysonEmail, dysonPassword string) error {
	myHotCoolLink := dyson.HotCoolLink{
		IP:               dysonIP,
		Port:             "1883",
		Serial:           dysonSerial,
		DysonAPIEmail:    dysonEmail,
		DysonAPIPassword: dysonPassword,
	}

	err := myHotCoolLink.Connect()
	if err != nil {
		return err
	}
	myHome.HotCool = myHotCoolLink
	log.WithFields(log.Fields{
		"dysonIP":       myHome.HotCool.IP,
		"dysonPort":     myHome.HotCool.Port,
		"dysonSerial":   myHome.HotCool.Serial,
		"dysonPassword": "******",
		"dysonEmail":    myHome.HotCool.DysonAPIEmail,
	}).Printf("new Dyson Hot Cold Link added")
	return nil
}

func (myHome *Home) AddTasmoto(OverrideTimeLength float64, URI string) error {
	myTasmoto := tasmoto.Switch{
		OverrideTimeLength: OverrideTimeLength,
		ManualOverride:     false,
		URI:                URI,
	}
	currentStatus, err := myTasmoto.CurrentStatus()
	if err != nil {
		return err
	}
	myTasmoto.AutomationStatus = *currentStatus
	log.WithFields(log.Fields{
		"URI":                myTasmoto.URI,
		"OverrideTimeLength": myTasmoto.OverrideTimeLength,
	}).Printf("new Tasmoto Switch added")
	return nil
}

func (myHome *Home) Monitor() error {

	myHome.HotCool.MonitorTemp()

	errChan := make(chan error)
	go ensureTemperature(&myHome.HVACSwitch, &myHome.HotCool, errChan, myHome.TemperatureChannel)
	for {
		setTempErr := <-errChan
		log.WithFields(log.Fields{
			"err": setTempErr.Error(),
		}).Printf("Error occurred when setting temperature")
	}
}

func ensureTemperature(allHVAC *[]tasmoto.Switch, homeSensor *dyson.HotCoolLink, errorChan chan error, tempChan chan float64) {
	tempPadding := 2.0
	desiredTemp := <-tempChan
	for {
		select {
		case newTemp := <-tempChan:
			desiredTemp = newTemp
			log.WithFields(log.Fields{
				"desiredTemp": desiredTemp,
			}).Printf("received new set temperature")
		default:
			for _, HVAC := range *allHVAC {
				err := HVAC.CheckManualOverride()
				if err != nil {
					errorChan <- err
				}
				if HVAC.ManualOverride && homeSensor.ClimateStatus.Data.Tact != "" {
					curTemp, err := strconv.ParseFloat(homeSensor.ClimateStatus.Data.Tact, 64)
					if err != nil {
						errorChan <- err
					}
					curTemp = (curTemp/10-273.15)*9/5 + 32
					log.WithFields(log.Fields{
						"curTemp": curTemp,
					}).Debugf("current temperature")
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
						//TODO USE URI FROM SWITCH
						_, err := client.Get("http://office-closet.oskoss.com/cm?cmnd=Power3%20On")
						if err != nil {
							errorChan <- err
						}
					} else if curTemp <= desiredTemp-tempPadding {
						timeout := time.Duration(5 * time.Second)
						client := &http.Client{
							Timeout: timeout,
						}
						//TODO USE URI FROM SWITCH
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
}

type Home struct {
	HotCool            dyson.HotCoolLink
	HVACSwitch         []tasmoto.Switch
	TemperatureChannel chan float64
}
