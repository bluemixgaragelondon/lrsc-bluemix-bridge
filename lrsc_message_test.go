package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LrscMessage", func() {
	Describe("toJson", func() {
		It("encodes to valid JSON", func() {
			m := lrscMessage{
				DeviceId: "AA-AA",
				Payload:  "test",
			}

			Expect(m.toJson()).To(Equal(`{"deveui":"AA-AA","pdu":"test"}`))
		})
	})
})
