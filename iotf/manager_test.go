package iotf

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("IotfManager", func() {
	Describe("extractCredentials", func() {
		It("extracts valid credentials", func() {
			vcapServices := `{"iotf-service":[{"name":"iotf","label":"iotf-service","tags":["internet_of_things","ibm_created"],"plan":"iotf-service-free","credentials":{"iotCredentialsIdentifier":"a2g6k39sl6r5","mqtt_host":"br2ybi.messaging.internetofthings.ibmcloud.com","mqtt_u_port":1883,"mqtt_s_port":8883,"base_uri":"https://internetofthings.ibmcloud.com:443/api/v0001","org":"br2ybi","apiKey":"a-br2ybi-y0tc7vicym","apiToken":"AJIpvsdJ!a__nqR(TK"}}]}`

			creds, err := extractCredentials(vcapServices)
			Expect(err).NotTo(HaveOccurred())
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
	})

	AfterEach(func() {
		close(eventsChannel)
		close(errorsChannel)
	})

	Describe("Connect", func() {
		It("calls connect on the broker", func() {
			Expect(iotfManager.Connect()).To(Succeed())
			Expect(mockBroker.connected).To(BeTrue())
		})
	})

	Describe("Loop", func() {
		It("publishes events on the broker", func() {
			go iotfManager.Loop()

			event := Event{Device: "device", Payload: "message"}
			eventRead := false
			select {
			case eventsChannel <- event:
				eventRead = true
			case <-time.After(time.Millisecond * 1):
				eventRead = false
			}

			Expect(eventRead).To(BeTrue())
			Expect(len(mockBroker.events)).To(Equal(1))
			Expect(mockBroker.events[0]).To(Equal(event))
		})

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

		Describe("device registration", func() {
			It("adds a device", func() {})
			It("doesn't add a device that has already been seen", func() {})
		})

		It("registers devices that have not yet been seen", func() {
			go iotfManager.Loop()

			event := Event{Device: "unseen", Payload: "message"}
			select {
			case eventsChannel <- event:
			case <-time.After(time.Millisecond * 1):
			}

			Expect(len(mockDeviceRegistrar.devices)).To(Equal(1))
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
	self.connected = true
	return nil
}

func (self *mockBroker) publishMessageFromDevice(event Event) {
	self.events = append(self.events, event)
}

type mockDeviceRegistrar struct {
	fail    bool
	devices map[string]struct{}
}

func newMockDeviceRegistrar() *mockDeviceRegistrar {
	return &mockDeviceRegistrar{devices: make(map[string]struct{})}
}

func (self *mockDeviceRegistrar) registerDevice(deviceId string) error {
	if self.fail {
		return errors.New("")
	}

	self.devices[deviceId] = struct{}{}
	return nil
}

func (self *mockDeviceRegistrar) deviceRegistered(deviceId string) bool {
	_, present := self.devices[deviceId]
	return present
}
