package iotf

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cromega/clogger"
	"github.com/bluemixgaragelondon/lrsc-bluemix-bridge/bridge"
	"github.com/bluemixgaragelondon/lrsc-bluemix-bridge/reporter"
	"github.com/bluemixgaragelondon/lrsc-bluemix-bridge/utils"
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

func NewIoTFManager(vcapServices string, commands chan<- bridge.Command, events <-chan Event, deviceType string) (*IoTFManager, error) {
	iotfCreds, err := extractCredentials(vcapServices)
	if err != nil {
		return nil, err
	}

	errChan := make(chan error)
	clientFactory := &mqttClientFactory{credentials: *iotfCreds, connectionLostHandler: func(err error) {
		logger.Error("IoTF connection lost handler called: " + err.Error())
		errChan <- errors.New("IoTF connection lost handler called: " + err.Error())
	}}

	broker := newIoTFBroker(iotfCreds, commands, errChan, deviceType, clientFactory)
	deviceRegistrar := newIotfHttpRegistrar(iotfCreds, deviceType)
	return &IoTFManager{broker: broker, deviceRegistrar: deviceRegistrar, events: events, errChan: errChan}, nil
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
