package main

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/reporter"
	"io"
)

var _ = Describe("LRSC Bridge", func() {
	It("validates handshake", func() {

		RegisterTestingT(GinkgoT())

		response := "some_handshake_response\n"
		Expect(validateHandshake(response)).To(Equal(true))
	})

	It("reconnects", func() {
		count := 0
		connectionAttempts := 0

		mockConn := &mockConnection{
			readFunc: func() (string, error) {
				if count == 0 {
					count += 1
					return "", errors.New("EOF")
				} else {
					return "{}\n", nil
				}
			},
			writeFunc: func(message string) error {
				if message == "JSON_000" {
					connectionAttempts += 1
				}
				return nil
			}}

		testDialer := &testDialer{conn: mockConn}
		lrscClient := &lrscConnection{dialer: testDialer}
		lrscClient.StatusReporter = reporter.New()
		lrscClient.inbound = make(chan lrscMessage)
		go runConnectionLoop("LRSC Client", lrscClient)

		messages := lrscClient.inbound
		<-messages
		Expect(connectionAttempts).To(Equal(2))
	})

	It("can receive a message", func() {
		mockConn := &mockConnection{
			readFunc: func() (string, error) {
				return `{"deveui": "id", "pdu": "data"}` + "\n", nil
			},
			writeFunc: func(string) error {
				return nil
			},
		}

		testDialer := &testDialer{conn: mockConn}
		lrscClient := &lrscConnection{dialer: testDialer}
		lrscClient.StatusReporter = reporter.New()
		lrscClient.inbound = make(chan lrscMessage)
		go runConnectionLoop("LRSC Client", lrscClient)

		messages := lrscClient.inbound
		Expect(<-messages).To(Equal(lrscMessage{Deveui: "id", Pdu: "data"}))
	})
	It("reports an error if connection fails", func() {
		failingDialer := &failingDialer{}
		lrscClient := &lrscConnection{dialer: failingDialer}
		lrscClient.StatusReporter = reporter.New()
		lrscClient.establish()

		Expect(lrscClient.Summary()).To(Equal(`{"CONNECTION":"FAILED"}`))
	})
})

type testDialer struct {
	conn io.ReadWriteCloser
}

func (self testDialer) dial() (io.ReadWriteCloser, error) {
	return self.conn, nil
}

func (self testDialer) endpoint() string {
	return "/dev/null"
}

type failingDialer struct {
}

func (self failingDialer) dial() (io.ReadWriteCloser, error) {
	return nil, errors.New("FAILED")
}

func (self failingDialer) endpoint() string {
	return "/dev/null"
}

type mockConnection struct {
	readFunc  func() (string, error)
	writeFunc func(string) error
	reporter.StatusReporter
}

func (self *mockConnection) Read(b []byte) (n int, err error) {
	response, err := self.readFunc()
	copy(b, response)
	return len(response), err
}

func (self *mockConnection) Write(b []byte) (n int, err error) {
	err = self.writeFunc(string(b))
	return len(b), err
}

func (self *mockConnection) Close() error {
	return nil
}
