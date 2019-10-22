package main

import (
	"os"
	"strconv"

	"github.com/oskoss/mi-casa/api"
	"github.com/oskoss/mi-casa/home"
	log "github.com/sirupsen/logrus"
)

func main() {
	configureLogging()
	//TODO Enable Config Setting via YAML
	//TODO Move config to separate function
	defaultDesiredTemp := 72.0
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.WithFields(log.Fields{
			"port": port,
		}).Printf("Web port not provided, using default port")
	}
	dysonEmail := os.Getenv("DYSON_API_EMAIL")
	dysonPassword := os.Getenv("DYSON_API_PASSWORD")
	dysonSerial := os.Getenv("DYSON_SERIAL_NUM")
	dysonIP := os.Getenv("DYSON_IP")
	userSetOverrideTimeString := os.Getenv("MANUAL_OVERRIDE_TIME_MINS")
	if userSetOverrideTimeString == "" {
		userSetOverrideTimeString = "60"
		log.WithFields(log.Fields{
			"userSetOverrideTimeString": userSetOverrideTimeString,
		}).Printf("MANUAL_OVERRIDE_TIME_MINS not provided, using default value")
	}
	userSetOverrideTime, err := strconv.ParseFloat(userSetOverrideTimeString, 64)
	if err != nil {
		log.Fatal(err)
	}
	tasmoto1 := os.Getenv("TASMOTO_URI1")
	if tasmoto1 == "" {
		tasmoto1 = "office-closet.oskoss.com"
		log.WithFields(log.Fields{
			"tasmoto1": tasmoto1,
		}).Printf("TASMOTO_URI1 not provided, using default value")
	}

	log.WithFields(log.Fields{
		"dysonIP":                   dysonIP,
		"dysonSerial":               dysonSerial,
		"dysonPassword":             "*****",
		"dysonEmail":                dysonEmail,
		"port":                      port,
		"userSetOverrideTimeString": userSetOverrideTimeString,
		"tasmoto1":                  tasmoto1,
	}).Printf("Configuration Successfully Applied")

	var myHome home.Home
	var temperatureChannel chan float64
	myHome.TemperatureChannel = temperatureChannel
	myHome.AddHotCoolLink(dysonIP, dysonSerial, dysonEmail, dysonPassword)
	myHome.AddTasmoto(userSetOverrideTime, tasmoto1)
	go myHome.Monitor()
	temperatureChannel <- defaultDesiredTemp
	api.Start(port, &myHome) // Blocking
}

func configureLogging() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)
	log.SetLevel(log.DebugLevel)
}
