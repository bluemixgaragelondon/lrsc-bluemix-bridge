package main

import (
	"errors"
	. "github.com/onsi/gomega"
	"io"
	"testing"
)

type TestDialer struct {
	conn io.ReadWriteCloser
}

func (self TestDialer) Dial() (io.ReadWriteCloser, error) {
	return self.conn, nil
}

func (self *TestDialer) Endpoint() string {
	return "/dev/null"
}

func TestValidateHandshake(t *testing.T) {
	RegisterTestingT(t)

	response := "some_handshake_response\n"
	Expect(validateHandshake(response)).To(Equal(true))
}

type MockConnection struct {
	readFunc  func() (string, error)
	writeFunc func(string) error
}

func (self *MockConnection) Read(b []byte) (n int, err error) {
	response, err := self.readFunc()
	copy(b, response)
	return len(response), err
}

func (self *MockConnection) Write(b []byte) (n int, err error) {
	err = self.writeFunc(string(b))
	return len(b), err
}

func (self *MockConnection) Close() error {
	return nil
}

func TestReceiveMessage(t *testing.T) {
	RegisterTestingT(t)

	mockConn := &MockConnection{
		readFunc: func() (string, error) {
			return `{"deveui": "id", "pdu": "data"}` + "\n", nil
		},
		writeFunc: func(string) error {
			return nil
		}}
	testDialer := &TestDialer{conn: mockConn}
	lrscConn := &LrscConnection{dialer: testDialer}
	messages := make(chan lrscMessage)

	lrscConn.StartListening(messages)
	Expect(<-messages).To(Equal(lrscMessage{Deveui: "id", Pdu: "data"}))
}

func TestReconnect(t *testing.T) {
	RegisterTestingT(t)

	count := 0
	connectionAttempts := 0

	mockConn := &MockConnection{
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

	testDialer := &TestDialer{conn: mockConn}
	lrscConn := &LrscConnection{dialer: testDialer}
	messages := make(chan lrscMessage)

	lrscConn.StartListening(messages)
	<-messages
	Expect(connectionAttempts).To(Equal(2))
}
