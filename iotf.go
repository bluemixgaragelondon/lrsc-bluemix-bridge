package main

import (
	"encoding/json"
	"errors"
	"fmt"
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type brokerConnection interface {
	connect() error
	publish(topic, message string)
}

type deviceRegistrar interface {
	registerDevice(deviceId string) (bool, error)
}

type mqttConnection struct {
	mqtt        *MQTT.MqttClient
	credentials *iotfCredentials
	deviceType  string
	err         chan error
}

type iotfRegistrar struct {
	credentials *iotfCredentials
	deviceType  string
}

type iotfConnection struct {
	devicesSeen  map[string]struct{}
	brokerClient brokerConnection
	registrar    deviceRegistrar
	err          chan error
	statusReporter
}

type iotfCredentials struct {
	User             string `json:"apiKey"`
	Password         string `json:"apiToken"`
	Org              string
	BaseUri          string `json:"base_uri"`
	MqttHost         string `json:"mqtt_host"`
	MqttSecurePort   int    `json:"mqtt_s_port"`
	MqttUnsecurePort int    `json:"mqtt_u_port"`
}

func (self *iotfConnection) initialise(creds *iotfCredentials, deviceType string) {
	self.devicesSeen = make(map[string]struct{})
	self.err = make(chan error)
	self.brokerClient = &mqttConnection{credentials: creds, deviceType: deviceType, err: self.err}
	self.registrar = &iotfRegistrar{credentials: creds, deviceType: deviceType}
}

func (self *iotfConnection) connect() error {
	err := self.brokerClient.connect()
	if err != nil {
		self.report("CONNECTION", err.Error())
	} else {
		self.report("CONNECTION", "OK")
	}
	return err
}

func (self *iotfConnection) error() chan error {
	return self.err
}

func (self *iotfConnection) loop() {
	//noop here as the mqtt library maintains it's own internal loop
}

func (self *iotfConnection) publish(device, message string) {
	if _, deviceFound := self.devicesSeen[device]; deviceFound == false {
		newDevice, err := self.registrar.registerDevice(device)
		if newDevice {
			self.devicesSeen[device] = struct{}{}
			self.report("DEVICES_SEEN", fmt.Sprintf("%v", len(self.devicesSeen)))
			self.report("LAST_REGISTRATION", "OK")
		}
		if err != nil {
			logger.Error("Could not register device: " + err.Error())
			self.report("LAST_REGISTRATION", err.Error())
		}
	}
	self.brokerClient.publish(device, message)
}

func (self *iotfRegistrar) registerDevice(device string) (bool, error) {
	registerUrl := fmt.Sprintf("%v/organizations/%v/devices", self.credentials.BaseUri, self.credentials.Org)
	body := strings.NewReader(fmt.Sprintf(`{"type": "%v", "id": "%v"}`, self.deviceType, device))
	request, err := http.NewRequest("POST", registerUrl, body)
	if err != nil {
		return false, err
	}

	request.SetBasicAuth(self.credentials.User, self.credentials.Password)
	request.Header.Add("Content-Type", "application/json")
	httpClient := http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		return false, err
	}
	responseBody, err := ioutil.ReadAll(response.Body)
	return deviceRegistered(response.StatusCode, responseBody)
}

func deviceRegistered(status int, body []byte) (bool, error) {
	switch status {
	case http.StatusForbidden:
		return false, errors.New("Did not autenticate successfully to IoTF")
	case http.StatusConflict:
		logger.Warning("Tried to register device that already exists: " + parseErrorFromIotf(body))
		return true, nil
	case http.StatusCreated:
		return true, nil
	default:
		return false, errors.New("Could not register device: " + parseErrorFromIotf(body))
	}
}

func parseErrorFromIotf(body []byte) string {
	parsedResponse := struct {
		Message string
	}{}

	err := json.Unmarshal(body, &parsedResponse)
	if err != nil {
		return "JSON parsing of response failed: " + err.Error()
	}
	return parsedResponse.Message
}

func (self *mqttConnection) connect() error {
	clientOpts := MQTT.NewClientOptions()
	clientOpts.AddBroker(fmt.Sprintf("tls://%v:%v", self.credentials.MqttHost, self.credentials.MqttSecurePort))
	clientOpts.SetClientId(fmt.Sprintf("a:%v:$v", self.credentials.Org, generateClientIdSuffix()))
	clientOpts.SetUsername(self.credentials.User)
	clientOpts.SetPassword(self.credentials.Password)

	clientOpts.SetOnConnectionLost(func(client *MQTT.MqttClient, mqttErr error) {
		logger.Error("IoTF connection lost handler called: " + mqttErr.Error())
		self.err <- errors.New("IoTF connection lost handler called: " + mqttErr.Error())
	})

	//MQTT.WARN = log.New(os.Stdout, "", 0)
	//MQTT.ERROR = log.New(os.Stdout, "", 0)
	//MQTT.DEBUG = log.New(os.Stdout, "", 0)

	self.mqtt = MQTT.NewClient(clientOpts)
	_, err := self.mqtt.Start()
	if err != nil {
		return errors.New("Could not establish MQTT connection: " + err.Error())
	}
	return nil
}

func (self *mqttConnection) publish(device, message string) {
	mqttMessage := MQTT.NewMessage([]byte(message))
	topic := fmt.Sprintf("iot-2/type/%v/id/%v/evt/TEST/fmt/json", self.deviceType, device)
	logger.Debug("Publishing '%v' to %v", message, topic)
	self.mqtt.PublishMessage(topic, mqttMessage)
}

func generateClientIdSuffix() string {
	rand.Seed(time.Now().UTC().UnixNano())
	suffix := rand.Intn(1000)
	return string(suffix)
}

func extractIotfCreds(services string) (*iotfCredentials, error) {
	data := struct {
		Services []struct {
			Credentials iotfCredentials
		} `json:"iotf-service"`
	}{}

	err := json.Unmarshal([]byte(services), &data)
	if err != nil {
		return nil, fmt.Errorf("Could not parse services JSON: %v", err)
	}

	if len(data.Services) == 0 {
		return nil, errors.New("Could not find any iotf-service instance bound")
	}

	return &data.Services[0].Credentials, nil
}
