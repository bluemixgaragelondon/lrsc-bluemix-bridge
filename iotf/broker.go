package iotf

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/bridge"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/mqtt"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/reporter"
	"regexp"
)

type broker interface {
	connect() error
	statusReporter() reporter.StatusReporter
	publishMessageFromDevice(Event)
}

type iotfBroker struct {
	client mqtt.Client
	reporter.StatusReporter
	commands      chan<- bridge.Command
	deviceType    string
	clientFactory clientFactory
	clientId      string
}

func newIoTFBroker(credentials *Credentials, commands chan<- bridge.Command, errChan chan<- error, deviceType string, clientFactory clientFactory) *iotfBroker {
	reporter := reporter.New()
	return &iotfBroker{commands: commands, StatusReporter: reporter, deviceType: deviceType, clientFactory: clientFactory, clientId: uuid.New()}
}

func (b *iotfBroker) connect() error {
	b.client = b.clientFactory.newClient(b.clientId)

	err := b.client.Start()
	if err != nil {
		b.Report("CONNECTION", err.Error())
		return err
	}

	b.Report("CONNECTION", "OK")

	logger.Info("Connected to MQTT")
	err = b.subscribeToCommandMessages(b.commands)
	if err != nil {
		b.Report("SUBSCRIPTION", err.Error())
		return err
	}
	b.Report("SUBSCRIPTION", "OK")
	return nil
}

func (self *iotfBroker) statusReporter() reporter.StatusReporter {
	return self.StatusReporter
}

func (self *iotfBroker) publishMessageFromDevice(event Event) {
	topic := fmt.Sprintf("iot-2/type/%v/id/%v/evt/TEST/fmt/json", self.deviceType, event.Device)
	logger.Debug("publishing event on topic %v: %v", topic, event)
	self.client.PublishMessage(topic, []byte(event.Payload))
}

func (self *iotfBroker) subscribeToCommandMessages(commands chan<- bridge.Command) error {
	topic := fmt.Sprintf("iot-2/type/%s/id/+/cmd/+/fmt/json", self.deviceType)
	return self.client.StartSubscription(topic, func(message mqtt.Message) {
		device := extractDeviceFromCommandTopic(message.Topic())
		command := bridge.Command{Device: device, Payload: string(message.Payload())}
		logger.Debug("received command message for %v", command.Device)
		commands <- command
	})
}

func extractDeviceFromCommandTopic(topic string) string {
	topicMatcher := regexp.MustCompile(`^iot-2/type/.*?/id/(.*?)/`)
	return topicMatcher.FindStringSubmatch(topic)[1]
}
