package iotf

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/mqtt"
)

var _ = Describe("IoTF Broker", func() {

	Describe("brokerConnection", func() {
		var (
			client     mockClient
			connection brokerConnection
		)

		BeforeEach(func() {
			client = NewMockClient()
			connection = brokerConnection{broker: &client}
		})

		Describe("PublishMessageFromDevice", func() {
			It("sends a message with the correct topic", func() {
				connection.publishMessageFromDevice(Event{Device: "foo", Payload: "message"})
				Expect(client.messages[0].Topic()).To(Equal("iot-2/type/LRSC/id/foo/evt/TEST/fmt/json"))
			})

			It("sends a message with the payload", func() {
				connection.publishMessageFromDevice(Event{Device: "foo", Payload: "message"})
				Expect(client.messages[0].Payload()).To(Equal([]byte("message")))
			})
		})

		Describe("SubscribeToCommandMessages", func() {
			It("puts received messages on a channel", func() {
				output := make(chan Command)
				go connection.subscribeToCommandMessages(output)
				Expect((<-output).Payload).To(Equal("command"))
			})

			It("extracts the device ID from the mqtt topic", func() {
				output := make(chan Command)
				go connection.subscribeToCommandMessages(output)
				Expect((<-output).Device).To(Equal("mydevice"))
			})
		})

	})

	Describe("extractDeviceFromCommandTopic", func() {
		It("returns the device ID", func() {
			topic := "iot-2/type/foo/id/devid/cmd/command/fmt/json"
			Expect(extractDeviceFromCommandTopic(topic)).To(Equal("devid"))
		})
	})

})

type mockClient struct {
	started  bool
	messages []mqtt.Message
}

func NewMockClient() mockClient {
	return mockClient{messages: []mqtt.Message{}}
}

func (self *mockClient) Start() error {
	self.started = true
	return nil
}

func (self *mockClient) PublishMessage(topic string, payload []byte) {
	self.messages = append(self.messages, message{topic, payload})
}

func (*mockClient) StartSubscription(_ string, callback func(message mqtt.Message)) error {
	topic := "iot-2/type/LRSC/id/mydevice/cmd/test/fmt/json"
	callback(message{topic: topic, payload: []byte("command")})
	return nil
}

type message struct {
	topic   string
	payload []byte
}

func (self message) Topic() string {
	return self.topic
}

func (self message) Payload() []byte {
	return self.payload
}
