package main

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestExtractIotfCreds(test *testing.T) {
	RegisterTestingT(test)

	vcapServices := `{"iotf-service":[{"name":"iotf","label":"iotf-service","tags":["internet_of_things","ibm_created"],"plan":"iotf-service-free","credentials":{"iotCredentialsIdentifier":"a2g6k39sl6r5","mqtt_host":"br2ybi.messaging.internetofthings.ibmcloud.com","mqtt_u_port":1883,"mqtt_s_port":8883,"base_uri":"https://internetofthings.ibmcloud.com:443/api/v0001","org":"br2ybi","apiKey":"a-br2ybi-y0tc7vicym","apiToken":"AJIpvsdJ!a__nqR(TK"}}]}`
	creds := extractIotfCreds(vcapServices)
	Expect(creds["user"]).To(Equal("a-br2ybi-y0tc7vicym"))
}
