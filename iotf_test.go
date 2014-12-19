package main

import (
	"errors"
	"github.com/cromega/clogger"
	. "github.com/onsi/gomega"
	"testing"
)

func init() {
	logger.SetLevel(clogger.Off)
}

func Test_IoTF_ValidCredentialsCanBeExtracted(test *testing.T) {
	RegisterTestingT(test)

	vcapServices := `{"iotf-service":[{"name":"iotf","label":"iotf-service","tags":["internet_of_things","ibm_created"],"plan":"iotf-service-free","credentials":{"iotCredentialsIdentifier":"a2g6k39sl6r5","mqtt_host":"br2ybi.messaging.internetofthings.ibmcloud.com","mqtt_u_port":1883,"mqtt_s_port":8883,"base_uri":"https://internetofthings.ibmcloud.com:443/api/v0001","org":"br2ybi","apiKey":"a-br2ybi-y0tc7vicym","apiToken":"AJIpvsdJ!a__nqR(TK"}}]}`

	creds, err := extractIotfCreds(vcapServices)
	Expect(err).NotTo(HaveOccurred())
	Expect(creds.User).To(Equal("a-br2ybi-y0tc7vicym"))
}

func Test_IoTF_EmptyMapVcapServicesProducesError(test *testing.T) {
	RegisterTestingT(test)

	vcapServices := "{}"

	_, err := extractIotfCreds(vcapServices)
	Expect(err).To(HaveOccurred())
}

func Test_IoTF_CompletelyEmptyVcapServicesProducesError(test *testing.T) {
	RegisterTestingT(test)

	vcapServices := ""

	_, err := extractIotfCreds(vcapServices)
	Expect(err).To(HaveOccurred())
}

func Test_IoTF_WrongServiceInVcapServicesProducesError(test *testing.T) {
	RegisterTestingT(test)

	vcapServices := `{"other-service":[{"credentials":{}}]}`

	_, err := extractIotfCreds(vcapServices)
	Expect(err).To(HaveOccurred())
}

func Test_IoTF_Publish_RegistersNewDevice(test *testing.T) {
	RegisterTestingT(test)

	client := createMockIotfClient()
	newDevice := "test"
	client.publish(newDevice, "Hello world")
	Expect(len(client.devicesSeen)).To(Equal(1))
}

func Test_IoTF_Publish_ReportsNewDevice(test *testing.T) {
	RegisterTestingT(test)

	client := createMockIotfClient()
	newDevice := "test"
	client.publish(newDevice, "Hello world")
	Expect(client.stats["DEVICES_SEEN"]).To(Equal("1"))
}

func Test_IoTF_Publish_DoesNotRegisterNewItemIfDeviceExist(test *testing.T) {
	RegisterTestingT(test)

	client := createMockIotfClient()
	client.devicesSeen["test"] = struct{}{}
	newDevice := "test"
	client.publish(newDevice, "Hello world")
	Expect(len(client.devicesSeen)).To(Equal(1))
}

func Test_IoTF_Publish_ReportsRegistrationFailure(test *testing.T) {
	RegisterTestingT(test)

	client := createMockIotfClient()
	client.registrar = &failingRegistrar{}
	newDevice := "test"
	client.publish(newDevice, "Hello world")
	Expect(client.stats["LAST_REGISTRATION"]).To(Equal("FAILED"))
}

func Test_IoTF_Connect_CreatesSuccessfulStatus(test *testing.T) {
	RegisterTestingT(test)

	client := createMockIotfClient()
	_ = client.connect()
	Expect(client.stats["CONNECTION"]).To(Equal("OK"))
}

func Test_IoTF_Connect_ReportsFailedConnection(test *testing.T) {
	RegisterTestingT(test)

	client := createMockIotfClient()
	client.brokerClient = &failingBroker{}
	_ = client.connect()
	Expect(client.stats["CONNECTION"]).To(Equal("FAILED"))
}

func createMockIotfClient() iotfConnection {
	devicesSeen := make(map[string]struct{})
	return iotfConnection{statusReporter: statusReporter{stats: make(map[string]string)}, devicesSeen: devicesSeen, brokerClient: &mockBroker{}, registrar: &mockRegistrar{}}
}

type mockBroker struct {
}

func (*mockBroker) connect() error {
	return nil
}

func (*mockBroker) publish(device, message string) {
}

type failingBroker struct {
}

func (*failingBroker) connect() error {
	return errors.New("FAILED")
}

func (*failingBroker) publish(device, message string) {
}

type mockRegistrar struct{}

func (*mockRegistrar) registerDevice(device string) (bool, error) {
	return true, nil
}

type failingRegistrar struct{}

func (*failingRegistrar) registerDevice(device string) (bool, error) {
	return false, errors.New("FAILED")
}
