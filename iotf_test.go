package main

import (
	. "github.com/onsi/gomega"
	"testing"
)

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
	client.Publish(newDevice, "Hello world")
	Expect(len(client.DevicesSeen)).To(Equal(1))
}

func Test_IoTF_Publish_ReportsNewDevice(test *testing.T) {
	RegisterTestingT(test)

	client := createMockIotfClient()
	newDevice := "test"
	client.Publish(newDevice, "Hello world")
	Expect(client.status["DEVICES_SEEN"]).To(Equal("1"))
}

func Test_IoTF_Publish_DoesNotRegisterNewItemIfDeviceExist(test *testing.T) {
	RegisterTestingT(test)

	client := createMockIotfClient()
	client.DevicesSeen["test"] = struct{}{}
	newDevice := "test"
	client.Publish(newDevice, "Hello world")
	Expect(len(client.DevicesSeen)).To(Equal(1))
}

func Test_IoTF_Connect_CreatesSuccessfulStatus(test *testing.T) {
	RegisterTestingT(test)

	client := createMockIotfClient()
	_ = client.Connect()
	Expect(client.status["CONNECTION"]).To(Equal("OK"))
}

func createMockIotfClient() iotfConnection {
	devicesSeen := make(map[string]struct{})
	return iotfConnection{StatusReporter: StatusReporter{status: make(map[string]string)}, DevicesSeen: devicesSeen, brokerClient: &mockBroker{}, registrar: &mockRegistrar{}}
}

type mockBroker struct {
}

func (*mockBroker) Connect() error {
	return nil
}

func (*mockBroker) Publish(device, message string) {
}

type mockRegistrar struct {
}

func (*mockRegistrar) RegisterDevice(device string) (bool, error) {
	return true, nil
}
