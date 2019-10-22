package dyson

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
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func (device *HotCoolLink) Connect() error {
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

	err := device.addDysonIntermediateCredentials()
	if err != nil {
		return err
	}

	err = device.addDysonAPIInfo()
	if err != nil {
		return err
	}

	client, err := device.connectToDevice()
	if err != nil {
		return err
	}
	device.MQTT = client
	return nil
}

func (device *HotCoolLink) MonitorTemp() {
	listenTopic := fmt.Sprintf("%s/%s/status/current", device.DysonAPIInfo.ProductType, device.DysonAPIInfo.Serial)
	stateTopic := fmt.Sprintf("%s/%s/command", device.DysonAPIInfo.ProductType, device.DysonAPIInfo.Serial)
	go device.listenAndUpdate(device.MQTT, listenTopic)
	go device.getState(device.MQTT, stateTopic)
}

type HotCoolLink struct {
	IP                           string
	Port                         string
	Serial                       string
	DysonAPIEmail                string
	DysonAPIPassword             string
	DecryptedDevicePassword      string
	DysonIntermediateCredentials Auth
	DysonAPIInfo                 Device
	ClimateStatus                Status
	MQTT                         mqtt.Client
}

type Auth struct {
	Account  string `json:"Account"`
	Password string `json:"Password"`
}

type Device struct {
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

type Status struct {
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

func (device *HotCoolLink) getState(client mqtt.Client, topic string) {
	timer := time.NewTicker(1 * time.Second)
	for range timer.C {
		client.Publish(topic, 0, false, "REQUEST-CURRENT-STATE")
	}
}

func (device *HotCoolLink) listenAndUpdate(client mqtt.Client, topic string) {
	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		var currentStatus Status
		message := msg.Payload()
		if err := json.Unmarshal(message, &currentStatus); err != nil {
			fmt.Printf("CurrentStatus %+v received from HotCoolLink %s is malformed", message, device.DysonAPIInfo.Serial)
		} else {
			device.ClimateStatus = currentStatus
		}
	})
}

func (device *HotCoolLink) connectToDevice() (mqtt.Client, error) {
	if device.DysonAPIInfo.Serial == "" {
		return nil, fmt.Errorf("device info from Dyson API has not been obtained yet")
	}
	if device.DysonAPIInfo.LocalCredentials == "" {
		return nil, fmt.Errorf("device info from Dyson API has not been obtained yet")
	}
	if device.IP == "" {
		return nil, fmt.Errorf("device IP not set")
	}
	if device.Port == "" {
		return nil, fmt.Errorf("device port not set")
	}
	if device.DecryptedDevicePassword == "" {
		return nil, fmt.Errorf("decrypted device password not set")
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
		return nil, err
	}

	return client, nil
}

func (device *HotCoolLink) addDysonIntermediateCredentials() error {
	if device.DysonAPIEmail == "" {
		return fmt.Errorf("Dyson Account Email has not be set")
	}
	if device.DysonAPIPassword == "" {
		return fmt.Errorf("Dyson Account Password has not be set")
	}
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

	resp, err := client.Post("https://api.cp.dyson.com/v1/userregistration/authenticate?country=US", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var authResp Auth

	if err := json.Unmarshal(respBytes, &authResp); err != nil {
		return err
	}
	device.DysonIntermediateCredentials = authResp
	fmt.Println("Added Dyson Intermediate Credentials")
	return nil
}

func (device *HotCoolLink) addDysonAPIInfo() error {

	if device.DysonIntermediateCredentials.Account == "" {
		return fmt.Errorf("Dyson intermediate credentials have not been obtained yet")
	}
	if device.DysonIntermediateCredentials.Password == "" {
		return fmt.Errorf("Dyson intermediate credentials have not been obtained yet")
	}

	timeout := time.Duration(5 * time.Second)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	req, err := http.NewRequest("GET", "https://api.cp.dyson.com/v1/provisioningservice/manifest", nil)
	req.SetBasicAuth(device.DysonIntermediateCredentials.Account, device.DysonIntermediateCredentials.Password)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var dysonAPIDevices []Device
	if err := json.Unmarshal(respBytes, &dysonAPIDevices); err != nil {
		return err
	}

	for _, dysonAPIDevice := range dysonAPIDevices {
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

func decryptPassword(encryptedPassword string) (string, error) {

	data, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", err
	}
	// Both key and initialization vector were found from https://libpurecoollink.readthedocs.io
	//1-32 HEX which act as the key
	key := []byte("\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f\x20")
	//16 HEX Zeros which act as the initialization vector
	iv := []byte("\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(data, data)
	unpaddedData, err := unpadByteArray(data)
	if err != nil {
		return "", err
	}
	var structuredData map[string]string
	err = json.Unmarshal(unpaddedData, &structuredData)
	if err != nil {
		return "", err
	}
	return structuredData["apPasswordHash"], nil
}

func unpadByteArray(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}
