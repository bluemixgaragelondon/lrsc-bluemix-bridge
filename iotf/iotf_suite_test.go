package iotf

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestIotf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IoTF Suite")
}
