package iotf

import (
	"errors"
	"fmt"
	"github.com/cromega/clogger"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/mqtt"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/reporter"
	"math/rand"
	"regexp"
	"time"
)

var logger clogger.Logger

type BrokerConnection struct {
	broker mqtt.Client
	reporter.StatusReporter
	events   <-chan Event
	commands chan<- Command
	errChan  chan error
}

const (
	deviceType = "LRSC"
)

func newClientOptions(credentials *Credentials, errChan chan<- error) mqtt.ClientOptions {
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

func NewBrokerConnection(credentials *Credentials, events <-chan Event, commands chan<- Command) *BrokerConnection {
	errChan := make(chan error)
	clientOptions := newClientOptions(credentials, errChan)
	broker := mqtt.NewPahoClient(clientOptions)
	return &BrokerConnection{broker: broker, events: events, commands: commands, errChan: errChan}
}

func (self *BrokerConnection) Connect() error {
	var err error
	err = self.broker.Start()
	if err != nil {
		return err
	}

	return self.subscribeToCommandMessages(self.commands)
}

func (self *BrokerConnection) Loop() {
	for event := range self.events {
		self.publishMessageFromDevice(event)
	}
}

func (self *BrokerConnection) Error() <-chan error {
	return self.errChan
}

func (self *BrokerConnection) publishMessageFromDevice(event Event) {
	topic := fmt.Sprintf("iot-2/type/%v/id/%v/evt/TEST/fmt/json", deviceType, event.Device)
	self.broker.PublishMessage(topic, []byte(event.Payload))
}

func (self *BrokerConnection) subscribeToCommandMessages(commands chan<- Command) error {
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
