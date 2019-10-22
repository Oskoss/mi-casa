package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/oskoss/mi-casa/home"
	log "github.com/sirupsen/logrus"
)

func handleV1HVACTemperature(myHome *home.Home) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		bodyString, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.WithFields(log.Fields{
				"err":      err,
				"req.Body": req.Body,
			}).Printf("could not read request")
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
		byteBody := []byte(bodyString)
		var HVACTempReq HVACSetTemp
		err = json.Unmarshal(byteBody, &HVACTempReq)
		if err != nil {
			log.WithFields(log.Fields{
				"err":      err,
				"req.Body": req.Body,
			}).Printf("could not un-marshal request")
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
		myHome.SetTemperature(HVACTempReq.Temperature)
		return
	}
}

type HVACSetTemp struct {
	Temperature float64 `json:"set_temperature"`
}
