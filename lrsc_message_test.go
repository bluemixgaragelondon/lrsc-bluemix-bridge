package main

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LrscMessage", func() {
	Describe("message modes", func() {
		It("confirmed", func() {
			Expect(messageModeConfirmed).To(Equal(lrscMessageMode(2)))
		})

		It("unconfirmed", func() {
			Expect(messageModeUnconfirmed).To(Equal(lrscMessageMode(0)))
		})
	})

	Describe("message types", func() {
		It("handshake", func() {
			Expect(messageTypeHandshake).To(Equal(lrscMessageType(0)))
		})

		It("upstream", func() {
			Expect(messageTypeUpstream).To(Equal(lrscMessageType(6)))
		})

		It("downstream", func() {
			Expect(messageTypeDownstream).To(Equal(lrscMessageType(7)))
		})
	})

	Describe("encoding to json", func() {
		It("results in valid LRSC JSON", func() {
			m := lrscMessage{
				DeviceGuid:       "AA-AA",
				Payload:          "test",
				UniqueSequenceNo: 658,
				Mode:             2,
				Timeout:          80,
				Port:             5,
			}

			mJson, _ := json.Marshal(m)

			Expect(mJson).To(MatchJSON(`{
				"deveui":"AA-AA",
				"pdu":"test",
				"seqno":658,
				"mode":2,
				"timeout":80,
				"port":5
			}`))
		})
	})
})
