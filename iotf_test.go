package main

import (
	"fmt"
	. "github.com/onsi/gomega"
	"testing"
)

func Test_IoTF_CredentialsCanBeExtracted(test *testing.T) {
	RegisterTestingT(test)

	vcapServices := `{"iotf-service":[{"name":"iotf","label":"iotf-service","tags":["internet_of_things","ibm_created"],"plan":"iotf-service-free","credentials":{"iotCredentialsIdentifier":"a2g6k39sl6r5","mqtt_host":"br2ybi.messaging.internetofthings.ibmcloud.com","mqtt_u_port":1883,"mqtt_s_port":8883,"base_uri":"https://internetofthings.ibmcloud.com:443/api/v0001","org":"br2ybi","apiKey":"a-br2ybi-y0tc7vicym","apiToken":"AJIpvsdJ!a__nqR(TK"}}]}`
	creds := extractIotfCreds(vcapServices)
	fmt.Printf("%v", creds)
	Expect(creds.User).To(Equal("a-br2ybi-y0tc7vicym"))
}

func Test_IoTF_Publish_RegistersNewDevice(test *testing.T) {
	RegisterTestingT(test)

	client := createMockIotfClient()
	newDevice := "test"
	client.Publish(newDevice, "Hello world")
	Expect(len(client.DevicesSeen)).To(Equal(1))
}

func Test_IoTF_Publish_DoesNotRegisterNewItemIfDeviceExist(test *testing.T) {
	RegisterTestingT(test)

	client := createMockIotfClient()
	client.DevicesSeen["test"] = struct{}{}
	newDevice := "test"
	client.Publish(newDevice, "Hello world")
	Expect(len(client.DevicesSeen)).To(Equal(1))
}

func createMockIotfClient() *iotfClient {
	devicesSeen := make(map[string]struct{})
	return &iotfClient{DevicesSeen: devicesSeen, broker: &mockBroker{}, registrar: &mockRegistrar{}}
}

type mockBroker struct {
}

func (*mockBroker) Publish(device, message string) {
}

type mockRegistrar struct {
}

func (*mockRegistrar) RegisterDevice(device string) (bool, error) {
	return true, nil
}
