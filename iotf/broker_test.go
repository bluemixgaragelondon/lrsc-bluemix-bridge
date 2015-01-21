package iotf

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/mqtt"
)

var _ = Describe("IoTF Broker", func() {
	var (
		client         *mockClient
		connection     *iotfBroker
		commandChannel chan Command
	)

	BeforeEach(func() {
		client = NewMockClient()
		reporter := &mockStatusReporter{}

		commandChannel = make(chan Command)

		connection = &iotfBroker{client: client, commands: commandChannel, StatusReporter: reporter}
	})

	AfterEach(func() {
		close(commandChannel)
	})

	Describe("connect", func() {
		Context("connection successful", func() {
			It("starts the broker", func() {
				connection.connect()
				Expect(client.started).To(Equal(true))
			})

			It("returns nil", func() {
				Expect(connection.connect()).ToNot(HaveOccurred())
			})

			It("subscribes to command messages", func() {
				connection.connect()
				client.fakePublish("command")

				Eventually(commandChannel).Should(Receive())
			})
		})

		Context("broker connection fails", func() {
			It("returns an error", func() {
				client.connectFail = true
				Expect(connection.connect()).To(HaveOccurred())
			})
		})

		Context("subscription fails", func() {
			It("returns an error", func() {
				client.subscribeFail = true
				Expect(connection.connect()).To(HaveOccurred())
			})
		})

	})

	Describe("publishMessageFromDevice", func() {
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
		BeforeEach(func() {
			client.started = true
		})

		It("puts received messages on a channel", func() {
			connection.subscribeToCommandMessages(commandChannel)
			client.fakePublish("command")
			Expect((<-commandChannel).Payload).To(Equal("command"))
		})

		It("extracts the device ID from the mqtt topic", func() {
			connection.subscribeToCommandMessages(commandChannel)
			client.fakePublish("command")
			Expect((<-commandChannel).Device).To(Equal("mydevice"))
		})
	})

})

var _ = Describe("extractDeviceFromCommandTopic", func() {
	It("returns the device ID", func() {
		topic := "iot-2/type/foo/id/devid/cmd/command/fmt/json"
		Expect(extractDeviceFromCommandTopic(topic)).To(Equal("devid"))
	})
})

type mockClient struct {
	connectFail          bool
	subscribeFail        bool
	started              bool
	messages             []mqtt.Message
	subscriptionCallback func(message mqtt.Message)
	topic                string
}

func NewMockClient() *mockClient {
	return &mockClient{messages: []mqtt.Message{}}
}

func (self *mockClient) Start() error {
	if self.connectFail {
		return errors.New("Could not start")
	}
	self.started = true
	return nil
}

func (self *mockClient) PublishMessage(topic string, payload []byte) {
	self.messages = append(self.messages, message{topic, payload})
}

func (self *mockClient) StartSubscription(_ string, callback func(message mqtt.Message)) error {
	if !self.started {
		return errors.New("subscription called when not connected")
	}
	if self.subscribeFail {
		return errors.New("an error")
	}
	self.subscriptionCallback = callback

	return nil
}

func (self *mockClient) fakePublish(messageData string) {
	topic := "iot-2/type/LRSC/id/mydevice/cmd/test/fmt/json"
	go self.subscriptionCallback(message{topic: topic, payload: []byte(messageData)})
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

type mockStatusReporter struct{}

func (self *mockStatusReporter) Report(key, value string) {}

func (self *mockStatusReporter) Summary() string {
	return ""
}
