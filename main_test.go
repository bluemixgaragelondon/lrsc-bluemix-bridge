package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/bridge"
)

var _ = Describe("Main - BREAK THIS DOWN ASAP", func() {
	Describe("converting IoTF commands to LRSC downstream messages", func() {
		lrscMessage := convertIotfCommandToLrscDownstreamMessage(bridge.Command{Device: "AA-AA", Payload: "payload"})

		It("message type is downstream", func() {
			Expect(lrscMessage.Type).To(Equal(messageTypeDownstream))
		})

		It("uses device id from IoTF command", func() {
			Expect(lrscMessage.DeviceGuid).To(Equal("AA-AA"))
		})

		It("uses payload from IoTF command", func() {
			Expect(lrscMessage.Payload).To(Equal("payload"))
		})
	})
})
