package iotf

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cromega/clogger"
)

var Logger clogger.Logger

type IoTFManager struct {
	broker Broker
	events <-chan Event
}

type Event struct {
	Device, Payload string
}

type Command struct {
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
	return &IoTFManager{broker: broker}, nil
}

func (self *IoTFManager) Connect() {
	self.broker.connect()
}

func (self *IoTFManager) Loop() {
	for event := range self.events {
		self.broker.publishMessageFromDevice(event)
	}
}

func (self *IoTFManager) Error() <-chan error {
	return nil
}

func extractCredentials(services string) (*Credentials, error) {
	data := struct {
		Services []struct {
			Credentials Credentials
		} `json:"iotf-service"`
	}{}

	err := json.Unmarshal([]byte(services), &data)
	if err != nil {
		Logger.Error("Could not parse services JSON: %v", err)
		return nil, fmt.Errorf("Could not parse services JSON: %v", err)
	}

	if len(data.Services) == 0 {
		Logger.Error("Could not find any iotf-service instance bound")
		return nil, errors.New("Could not find any iotf-service instance bound")
	}

	return &data.Services[0].Credentials, nil
}
