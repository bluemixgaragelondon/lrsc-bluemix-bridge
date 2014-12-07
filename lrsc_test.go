package main

import (
	. "github.com/onsi/gomega"
	"io"
	"testing"
)

type TestDialer struct {
}

func (self *TestDialer) Dial() (io.ReadWriteCloser, error) {
	return &MockConnection{}, nil
}

func (self *TestDialer) Endpoint() string {
	return "/dev/null"
}

type MockConnection struct {
}

func (self *MockConnection) Read(b []byte) (n int, err error) {
	response := []byte("response\n")
	copy(b, response)
	return len(response), nil
}

func (self *MockConnection) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (self *MockConnection) Close() error {
	return nil
}

func TestHandshake(t *testing.T) {
	RegisterTestingT(t)

	testDialer := &TestDialer{}
	lrscConn := &LrscConnection{dialer: testDialer}
	messages := make(chan string)
	lrscConn.StartListening(messages)
	message := <-messages

	Expect(message).To(Equal("response"))

}
