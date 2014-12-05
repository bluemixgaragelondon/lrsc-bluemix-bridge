package main

import (
	"fmt"
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
	//fmt.Println("Read called")
	response := []byte("response\n")
	fmt.Printf("Length: %v", len(response))
	fmt.Println(len(b))
	b = append(response, b...)
	//copy(b[:], response)
	return len(b), nil
}

func (self *MockConnection) Write(b []byte) (n int, err error) {
	fmt.Printf("Write called with '%v'\n", string(b))
	return len(b), nil
}

func (self *MockConnection) Close() error {
	fmt.Printf("Close called")
	return nil
}

func TestHandshake(t *testing.T) {
	RegisterTestingT(t)

	testDialer := &TestDialer{}
	lrscConn := &LrscConnection{dialer: testDialer}
	messages := make(chan string)
	lrscConn.StartListening(messages)
	message := <-messages
	fmt.Printf("%v", []byte(message))
	Expect(message).To(Equal("response"))

}
