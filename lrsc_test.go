package main

import (
	. "github.com/onsi/gomega"
	"os"
	"testing"
)

func TestLrscConnectionSucceeds(test *testing.T) {
	cert := readCertificate(os.Getenv("CLIENT_CERT"))
	key := readCertificate(os.Getenv("CLIENT_KEY"))

	lrscConn := CreateLrscConnection("dev.lrsc.ch", "55055", cert, key)

	lrscConn.Connect()
}

func TestLrscConnectionFails(test *testing.T) {
	cert := readCertificate(os.Getenv("CLIENT_CERT"))
	key := readCertificate(os.Getenv("CLIENT_KEY"))

	lrscConn := CreateLrscConnection("foobar", "55055", cert, key)
	Expect(func() { lrscConn.Connect() }).To(Panic())
}
