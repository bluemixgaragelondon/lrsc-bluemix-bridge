package iotf

import (
	"github.com/cromega/clogger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestIotf(t *testing.T) {
	RegisterFailHandler(Fail)

	logger.SetLevel(clogger.Off)
	RunSpecs(t, "IoTF Suite")
}
