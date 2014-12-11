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

type BrokerClient interface {
	Publish(topic, message string)
}

type DeviceRegistrar interface {
	RegisterDevice(deviceId string) (bool, error)
}

type mqttClient struct {
	mqtt        *MQTT.MqttClient
	credentials iotfCredentials
	deviceType  string
}

type iotfRegistrar struct {
	credentials iotfCredentials
	deviceType  string
}

type iotfClient struct {
	DevicesSeen map[string]struct{}
	broker      BrokerClient
	registrar   DeviceRegistrar
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

func CreateIotfClient(creds iotfCredentials, deviceType string) (*iotfClient, error) {

	clientOpts := MQTT.NewClientOptions()
	clientOpts.AddBroker(fmt.Sprintf("tls://%v:%v", creds.MqttHost, creds.MqttSecurePort))
	clientOpts.SetClientId(fmt.Sprintf("a:%v:$v", creds.Org, generateClientIdSuffix()))
	clientOpts.SetUsername(creds.User)
	clientOpts.SetPassword(creds.Password)

	clientOpts.SetOnConnectionLost(func(client *MQTT.MqttClient, err error) {
		logger.Error("IoTF connection lost handler called: " + err.Error())
	})

	//MQTT.WARN = log.New(os.Stdout, "", 0)
	//MQTT.ERROR = log.New(os.Stdout, "", 0)
	//MQTT.DEBUG = log.New(os.Stdout, "", 0)

	mqtt := MQTT.NewClient(clientOpts)
	_, err := mqtt.Start()
	if err != nil {
		return nil, errors.New("Could not establish MQTT connection: " + err.Error())
	}

	devicesSeen := make(map[string]struct{})

	broker := &mqttClient{credentials: creds, deviceType: deviceType, mqtt: mqtt}
	registrar := &iotfRegistrar{credentials: creds, deviceType: deviceType}
	return &iotfClient{DevicesSeen: devicesSeen, broker: broker, registrar: registrar}, nil
}

func (self *iotfClient) Publish(device, message string) {
	if _, deviceFound := self.DevicesSeen[device]; deviceFound == false {
		newDevice, err := self.registrar.RegisterDevice(device)
		if newDevice {
			self.DevicesSeen[device] = struct{}{}
		}
		if err != nil {
			logger.Error("Could not register device: " + err.Error())
		}
	}
	self.broker.Publish(device, message)
}

func (self *iotfRegistrar) RegisterDevice(device string) (bool, error) {
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

func (self *mqttClient) Publish(device, message string) {
	mqttMessage := MQTT.NewMessage([]byte(message))
	topic := fmt.Sprintf("iot-2/type/%v/id/%v/evt/TEST/fmt/json", self.deviceType, device)
	logger.Info("Publishing '%v' to %v", message, topic)
	self.mqtt.PublishMessage(topic, mqttMessage)
}

func generateClientIdSuffix() string {
	rand.Seed(time.Now().UTC().UnixNano())
	suffix := rand.Intn(1000)
	return string(suffix)
}

func extractIotfCreds(services string) iotfCredentials {
	data := struct {
		Services []struct {
			Credentials iotfCredentials
		} `json:"iotf-service"`
	}{}

	err := json.Unmarshal([]byte(services), &data)
	if err != nil {
		logger.Error(fmt.Sprintf("%v (probably missing configuration)", err))
	}

	return data.Services[0].Credentials
}
