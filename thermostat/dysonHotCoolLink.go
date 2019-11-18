package thermostat

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

type DysonHotCoolLink struct {
	Name                         string `yaml:"name"`
	IP                           string `yaml:"ip"`
	Port                         string `yaml:"port"`
	Serial                       string `yaml:"serialNumber"`
	DysonAPIEmail                string `yaml:"dysonAPIEmail"`
	DysonAPIPassword             string `yaml:"dysonAPIPassword"`
	DysonAPIEndpoint             string `yaml:"dysonAPIEndpoint,omitempty"`
	DecryptedDevicePassword      string
	DysonIntermediateCredentials DysonAuth
	DysonAPIInfo                 DysonAPIInfo
	ClimateStatus                DysonHotCoolLinkStatus
	MQTT                         mqtt.Client
}

type DysonAuth struct {
	Account  string `json:"Account"`
	Password string `json:"Password"`
}

type DysonAPIInfo struct {
	Active              bool   `json:"Active"`
	Serial              string `json:"Serial"`
	Name                string `json:"Name"`
	ScaleUnit           string `json:"ScaleUnit"`
	Version             string `json:"Version"`
	LocalCredentials    string `json:"LocalCredentials"`
	AutoUpdate          bool   `json:"AutoUpdate"`
	NewVersionAvailable bool   `json:"NewVersionAvailable"`
	ProductType         string `json:"ProductType"`
}

type DysonHotCoolLinkStatus struct {
	Msg  string    `json:"msg"`
	Time time.Time `json:"time"`
	Data struct {
		Tact string `json:"tact"`
		Hact string `json:"hact"`
		Pact string `json:"pact"`
		Vact string `json:"vact"`
		Sltm string `json:"sltm"`
	} `json:"data"`
}

func (device *DysonHotCoolLink) CurrentTemp() (temp *float64, err error) {
	if device.ClimateStatus.Data.Tact == "" {
		return nil, fmt.Errorf("Temperature Not Retrieved Yet")
	}
	curTemp, err := strconv.ParseFloat(device.ClimateStatus.Data.Tact, 64)
	if err != nil {
		return nil, err
	}
	curTemp = (curTemp/10-273.15)*9/5 + 32
	log.WithFields(log.Fields{
		"curTemp": curTemp,
	}).Debugf("current temperature")
	return &curTemp, nil
}

func (device *DysonHotCoolLink) Connect() (err error) {
	if device.DysonAPIEmail == "" {
		return fmt.Errorf("HotCoolLink device DysonAPIEmail not set")
	}
	if device.DysonAPIPassword == "" {
		return fmt.Errorf("HotCoolLink device DysonAPIPassword not set")
	}
	if device.IP == "" {
		return fmt.Errorf("HotCoolLink device IP not set")
	}
	if device.Port == "" {
		return fmt.Errorf("HotCoolLink device port not set")
	}
	if device.Serial == "" {
		return fmt.Errorf("HotCoolLink device Serial not set")
	}

	err = device.addDysonIntermediateCredentials()
	if err != nil {
		return err
	}
	err = device.addDysonAPIInfo()
	if err != nil {
		return err
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%s", device.IP, device.Port))
	opts.SetUsername(device.DysonAPIInfo.Serial)
	opts.SetPassword(device.DecryptedDevicePassword)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		return err
	}

	device.MQTT = client
	listenTopic := fmt.Sprintf("%s/%s/status/current", device.DysonAPIInfo.ProductType, device.DysonAPIInfo.Serial)
	stateTopic := fmt.Sprintf("%s/%s/command", device.DysonAPIInfo.ProductType, device.DysonAPIInfo.Serial)
	go device.SubscribeTemp(device.MQTT, listenTopic)
	go device.RequestTemp(device.MQTT, stateTopic)
	return nil
}

func (device *DysonHotCoolLink) RequestTemp(client mqtt.Client, topic string) {
	timer := time.NewTicker(1 * time.Second)
	for range timer.C {
		client.Publish(topic, 0, false, "REQUEST-CURRENT-STATE")
	}
}

func (device *DysonHotCoolLink) SubscribeTemp(client mqtt.Client, topic string) {
	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		var currentStatus DysonHotCoolLinkStatus
		message := msg.Payload()
		if err := json.Unmarshal(message, &currentStatus); err != nil {
			fmt.Printf("CurrentStatus %+v received from HotCoolLink %s is malformed", message, device.DysonAPIInfo.Serial)
		} else {
			device.ClimateStatus = currentStatus
		}
	})
}

func (device *DysonHotCoolLink) addDysonIntermediateCredentials() (err error) {

	timeout := time.Duration(5 * time.Second)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	requestBody, err := json.Marshal(map[string]string{
		"Email":    device.DysonAPIEmail,
		"Password": device.DysonAPIPassword,
	})

	if err != nil {
		return err
	}
	if device.DysonAPIEndpoint == "" {
		device.DysonAPIEndpoint = "https://api.cp.dyson.com"
	}
	resp, err := client.Post(
		fmt.Sprintf("%s/v1/userregistration/authenticate?country=US",
			device.DysonAPIEndpoint),
		"application/json",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to login to dyson API check username %s and password", device.DysonAPIEmail)
	}

	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var authResp DysonAuth

	err = json.Unmarshal(respBytes, &authResp)
	if err != nil {
		return err
	}
	device.DysonIntermediateCredentials = authResp
	fmt.Println("Added Dyson Intermediate Credentials")
	return nil
}

func (device *DysonHotCoolLink) addDysonAPIInfo() (err error) {

	timeout := time.Duration(5 * time.Second)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/v1/provisioningservice/manifest", device.DysonAPIEndpoint),
		nil,
	)
	req.SetBasicAuth(
		device.DysonIntermediateCredentials.Account,
		device.DysonIntermediateCredentials.Password,
	)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to login to dyson API with intermediate credentials Account: %s, Password: %s", device.DysonIntermediateCredentials.Account, device.DysonIntermediateCredentials.Password)
	}
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var dysonAPIDeviceInfo []DysonAPIInfo
	if err := json.Unmarshal(respBytes, &dysonAPIDeviceInfo); err != nil {
		return err
	}

	for _, dysonAPIDevice := range dysonAPIDeviceInfo {
		if dysonAPIDevice.Serial == device.Serial {
			device.DysonAPIInfo = dysonAPIDevice
			decryptedDevicePassword, err := decryptPassword(device.DysonAPIInfo.LocalCredentials)
			if err != nil {
				return err
			}
			device.DecryptedDevicePassword = decryptedDevicePassword
			fmt.Println("Added Dyson API Info")
			return nil
		}
	}
	return fmt.Errorf("Device with serial %s not found in dyson account with email %s", device.Serial, device.DysonAPIEmail)
}

func decryptPassword(encryptedPassword string) (decryptedPass string, err error) {

	data, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return decryptedPass, err
	}
	// Both key and initialization vector were found from https://libpurecoollink.readthedocs.io
	//1-32 HEX which act as the key
	key := []byte("\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f\x20")
	//16 HEX Zeros which act as the initialization vector
	iv := []byte("\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")

	block, err := aes.NewCipher(key)
	if err != nil {
		return decryptedPass, err
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(data, data)
	unpaddedData, err := unpadByteArray(data)
	if err != nil {
		return decryptedPass, err
	}
	var structuredData map[string]string
	err = json.Unmarshal(unpaddedData, &structuredData)
	if err != nil {
		return decryptedPass, err
	}
	decryptedPass = structuredData["apPasswordHash"]
	return decryptedPass, nil
}

func unpadByteArray(src []byte) ([]byte, error) {
	length := len(src)
	if length <= 1 {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}
