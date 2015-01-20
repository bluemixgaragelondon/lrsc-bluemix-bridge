package main

import (
	"github.com/cromega/clogger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLrscBridge(t *testing.T) {
	RegisterFailHandler(Fail)

	logger.SetLevel(clogger.Off)
	RunSpecs(t, "LrscBridge Suite")
}
