package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLrscBridge(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "LrscBridge Suite")
}
