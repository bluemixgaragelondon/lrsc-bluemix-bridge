package iotf

import (
	"errors"
	"fmt"
	"github.com/cromega/clogger"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/mqtt"
	"math/rand"
	"regexp"
	"time"
)

var logger clogger.Logger

type brokerConnection struct {
	broker   mqtt.Client
	events   <-chan Event
	commands chan<- Command
	errChan  chan<- error
}

const (
	deviceType = "LRSC"
)

func newClientOptions(credentials iotfCredentials, errChan chan<- error) mqtt.ClientOptions {
	return mqtt.ClientOptions{
		Broker:   fmt.Sprintf("tls://%v:%v", credentials.MqttHost, credentials.MqttSecurePort),
		ClientId: fmt.Sprintf("a:%v:$v", credentials.Org, generateClientIdSuffix()),
		Username: credentials.User,
		Password: credentials.Password,
		OnConnectionLost: func(err error) {
			logger.Error("IoTF connection lost handler called: " + err.Error())
			errChan <- errors.New("IoTF connection lost handler called: " + err.Error())
		},
	}
}

func generateClientIdSuffix() string {
	rand.Seed(time.Now().UTC().UnixNano())
	suffix := rand.Intn(1000)
	return string(suffix)
}

func newBrokerConnection(credentials iotfCredentials, events <-chan Event, commands chan<- Command, errChan chan<- error) brokerConnection {
	clientOptions := newClientOptions(credentials, errChan)
	broker := mqtt.NewPahoClient(clientOptions)
	return brokerConnection{broker: broker, events: events, commands: commands, errChan: errChan}
}

func (self *brokerConnection) connect() error {
	var err error
	err = self.broker.Start()
	if err != nil {
		return err
	}

	return self.subscribeToCommandMessages(self.commands)
}

func (self *brokerConnection) run() {
	for event := range self.events {
		self.publishMessageFromDevice(event)
	}
}

func (self *brokerConnection) publishMessageFromDevice(event Event) {
	topic := fmt.Sprintf("iot-2/type/%v/id/%v/evt/TEST/fmt/json", deviceType, event.Device)
	self.broker.PublishMessage(topic, []byte(event.Payload))
}

func (self *brokerConnection) subscribeToCommandMessages(commands chan<- Command) error {
	topic := fmt.Sprintf("iot-2/type/%s/id/+/cmd/+/fmt/json", deviceType)
	return self.broker.StartSubscription(topic, func(message mqtt.Message) {
		device := extractDeviceFromCommandTopic(message.Topic())
		command := Command{Device: device, Payload: string(message.Payload())}
		commands <- command
	})
}

func extractDeviceFromCommandTopic(topic string) string {
	topicMatcher := regexp.MustCompile(`^iot-2/type/.*?/id/(.*?)/`)
	return topicMatcher.FindStringSubmatch(topic)[1]
}
