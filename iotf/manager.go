package iotf

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cromega/clogger"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/reporter"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/utils"
)

var logger clogger.Logger

func init() {
	logger = utils.CreateLogger()
}

type IoTFManager struct {
	broker          broker
	deviceRegistrar deviceRegistrar
	events          <-chan Event
	errChan         chan error
}

type Event struct {
	Device, Payload string
}

type Credentials struct {
	User             string `json:"apiKey"`
	Password         string `json:"apiToken"`
	Org              string
	BaseUri          string `json:"base_uri"`
	MqttHost         string `json:"mqtt_host"`
	MqttSecurePort   int    `json:"mqtt_s_port"`
	MqttUnsecurePort int    `json:"mqtt_u_port"`
}

func NewIoTFManager(vcapServices string, commands chan<- Command, events <-chan Event) (*IoTFManager, error) {
	iotfCreds, err := extractCredentials(vcapServices)
	if err != nil {
		return nil, err
	}

	errChan := make(chan error)
	broker := newIoTFBroker(iotfCreds, commands, errChan)
	deviceRegistrar := newIotfHttpRegistrar(iotfCreds)
	return &IoTFManager{broker: broker, deviceRegistrar: deviceRegistrar, errChan: errChan}, nil
}

func (self *IoTFManager) Connect() error {
	return self.broker.connect()
}

func (self *IoTFManager) Loop() {
	for event := range self.events {
		self.deviceRegistrar.registerDevice(event.Device)
		self.broker.publishMessageFromDevice(event)
	}
}

func (self *IoTFManager) Error() <-chan error {
	return self.errChan
}

func (self *IoTFManager) StatusReporter() reporter.StatusReporter {
	return self.broker.statusReporter()
}

func extractCredentials(services string) (*Credentials, error) {
	data := struct {
		Services []struct {
			Credentials Credentials
		} `json:"iotf-service"`
	}{}

	err := json.Unmarshal([]byte(services), &data)
	if err != nil {
		logger.Error("Could not parse services JSON: %v", err)
		return nil, fmt.Errorf("Could not parse services JSON: %v", err)
	}

	if len(data.Services) == 0 {
		logger.Error("Could not find any iotf-service instance bound")
		return nil, errors.New("Could not find any iotf-service instance bound")
	}

	return &data.Services[0].Credentials, nil
}
