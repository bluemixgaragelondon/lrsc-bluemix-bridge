package iotf

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/reporter"
	"time"
)

var (
	callOrder []string
)

var _ = Describe("IotfManager", func() {
	Describe("extractCredentials", func() {
		It("extracts valid credentials", func() {
			vcapServices := `{"iotf-service":[{"name":"iotf","label":"iotf-service","tags":["internet_of_things","ibm_created"],"plan":"iotf-service-free","credentials":{"iotCredentialsIdentifier":"a2g6k39sl6r5","mqtt_host":"br2ybi.messaging.internetofthings.ibmcloud.com","mqtt_u_port":1883,"mqtt_s_port":8883,"base_uri":"https://internetofthings.ibmcloud.com:443/api/v0001","org":"br2ybi","apiKey":"a-br2ybi-y0tc7vicym","apiToken":"AJIpvsdJ!a__nqR(TK"}}]}`

			creds, _ := extractCredentials(vcapServices)
			Expect(creds.User).To(Equal("a-br2ybi-y0tc7vicym"))
		})

		It("errors with empty VCAP_SERVICES", func() {
			vcapServices := "{}"

			_, err := extractCredentials(vcapServices)
			Expect(err).To(HaveOccurred())
		})

		It("errors with empty string", func() {
			vcapServices := ""

			_, err := extractCredentials(vcapServices)
			Expect(err).To(HaveOccurred())
		})

	})

	var (
		iotfManager         *IoTFManager
		mockBroker          *mockBroker
		mockDeviceRegistrar *mockDeviceRegistrar
		eventsChannel       chan Event
		errorsChannel       chan error
	)
	BeforeEach(func() {
		eventsChannel = make(chan Event)
		errorsChannel = make(chan error)
		mockBroker = newMockBroker()
		mockDeviceRegistrar = newMockDeviceRegistrar()
		iotfManager = &IoTFManager{broker: mockBroker, deviceRegistrar: mockDeviceRegistrar,
			events: eventsChannel, errChan: errorsChannel}
		callOrder = make([]string, 0)
	})

	AfterEach(func() {
		close(eventsChannel)
		close(errorsChannel)
	})

	Describe("Connect", func() {
		It("succeeds when broker connects", func() {
			mockBroker.connected = true

			Expect(iotfManager.Connect()).To(Succeed())
		})

		It("fails when broker does not connect", func() {
			mockBroker.connected = false

			Expect(iotfManager.Connect()).ToNot(Succeed())
		})
	})

	Describe("Loop", func() {
		It("loops", func() {
			go iotfManager.Loop()

			event := Event{Device: "device", Payload: "message"}

			for i := 0; i < 5; i++ {
				select {
				case eventsChannel <- event:
				case <-time.After(time.Millisecond * 1):
				}
			}

			Expect(len(mockBroker.events)).To(Equal(5))
		})

		Context("when an event is received", func() {
			It("publishes to the broker", func() {
				event := Event{Device: "device", Payload: "message"}

				go iotfManager.Loop()
				select {
				case eventsChannel <- event:
				case <-time.After(time.Millisecond * 1):
				}

				Expect(mockBroker.events).To(Equal([]Event{event}))
			})
		})

		Describe("device registration", func() {
			It("registers devices that have not yet been seen", func() {
				go iotfManager.Loop()

				event := Event{Device: "unseen", Payload: "message"}
				select {
				case eventsChannel <- event:
				case <-time.After(time.Millisecond * 1):
				}

				_, devicePresent := mockDeviceRegistrar.devices["unseen"]
				Expect(devicePresent).To(BeTrue())
			})

			It("registers the device before it publishes the message", func() {
				go iotfManager.Loop()

				event := Event{Device: "unseen", Payload: "message"}
				select {
				case eventsChannel <- event:
				case <-time.After(time.Millisecond * 1):
				}

				Expect(callOrder).To(Equal([]string{"registerDevice", "publishMessageFromDevice"}))
			})
		})

	})
	Describe("Error", func() {
		It("returns the managers read-only error channel", func() {
			var errChan <-chan error = iotfManager.errChan
			Expect(iotfManager.Error()).To(Equal(errChan))
		})
	})
})

type mockBroker struct {
	connected bool
	events    []Event
}

func newMockBroker() *mockBroker {
	events := make([]Event, 0)
	return &mockBroker{events: events}
}

func (self *mockBroker) connect() error {
	if !self.connected {
		return errors.New("failed to connect")
	}
	return nil
}

func (self *mockBroker) statusReporter() reporter.StatusReporter {
	return nil
}

func (self *mockBroker) publishMessageFromDevice(event Event) {
	callOrder = append(callOrder, "publishMessageFromDevice")
	self.events = append(self.events, event)
}

type mockDeviceRegistrar struct {
	connect bool
	devices map[string]struct{}
}

func newMockDeviceRegistrar() *mockDeviceRegistrar {
	return &mockDeviceRegistrar{devices: make(map[string]struct{})}
}

func (self *mockDeviceRegistrar) registerDevice(deviceId string) error {
	callOrder = append(callOrder, "registerDevice")
	if self.connect {
		return errors.New("")
	}

	self.devices[deviceId] = struct{}{}
	return nil
}

func (self *mockDeviceRegistrar) deviceRegistered(deviceId string) bool {
	_, present := self.devices[deviceId]
	return present
}
