package main

import (
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"testing"
)

func TestLrscConnectionSucceeds(test *testing.T) {
	RegisterTestingT(test)

	cert, err := ioutil.ReadFile(os.Getenv("CLIENT_CERT"))
	Expect(err).To(BeNil())

	key, err := ioutil.ReadFile(os.Getenv("CLIENT_KEY"))
	Expect(err).To(BeNil())

	lrscConn, err := CreateLrscConnection("dev.lrsc.ch", "55055", cert, key)
	Expect(err).To(BeNil())

	err = lrscConn.Connect()
	Expect(err).To(BeNil())
}

func TestLrscConnectionFails(test *testing.T) {
	RegisterTestingT(test)

	cert, err := ioutil.ReadFile(os.Getenv("CLIENT_CERT"))
	Expect(err).To(BeNil())

	key, err := ioutil.ReadFile(os.Getenv("CLIENT_KEY"))
	Expect(err).To(BeNil())

	lrscConn, err := CreateLrscConnection("foobar", "55055", cert, key)
	Expect(err).To(BeNil())

	err = lrscConn.Connect()
	Expect(err).To(HaveOccurred())
}
