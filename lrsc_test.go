package main

import (
	"errors"
	. "github.com/onsi/gomega"
	"io"
	"testing"
)

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

func TestValidateHandshake(t *testing.T) {
	RegisterTestingT(t)

	response := "some_handshake_response\n"
	Expect(validateHandshake(response)).To(Equal(true))
}

type mockConnection struct {
	readFunc  func() (string, error)
	writeFunc func(string) error
	statusReporter
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

func Test_LRSC_CanReceiveMessage(t *testing.T) {
	RegisterTestingT(t)

	mockConn := &mockConnection{
		readFunc: func() (string, error) {
			return `{"deveui": "id", "pdu": "data"}` + "\n", nil
		},
		writeFunc: func(string) error {
			return nil
		}}

	testDialer := &testDialer{conn: mockConn}
	lrscClient := &lrscConnection{dialer: testDialer}
	lrscClient.stats = make(map[string]string)
	lrscClient.inbound = make(chan lrscMessage)
	go runConnectionLoop("LRSC Client", lrscClient)

	messages := lrscClient.inbound
	Expect(<-messages).To(Equal(lrscMessage{Deveui: "id", Pdu: "data"}))
}

func Test_LRSC_Reconnects(t *testing.T) {
	RegisterTestingT(t)

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
	lrscClient.stats = make(map[string]string)
	lrscClient.inbound = make(chan lrscMessage)
	go runConnectionLoop("LRSC Client", lrscClient)

	messages := lrscClient.inbound
	<-messages
	Expect(connectionAttempts).To(Equal(2))
}

func Test_LRSC_ReportsErrorIfConnectionFails(t *testing.T) {
	RegisterTestingT(t)

	failingDialer := &failingDialer{}
	lrscClient := &lrscConnection{dialer: failingDialer}
	lrscClient.stats = make(map[string]string)
	lrscClient.establish()

	Expect(lrscClient.stats["CONNECTION"]).To(Equal("FAILED"))
}
