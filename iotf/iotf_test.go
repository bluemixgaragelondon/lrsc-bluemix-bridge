package iotf

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Iotf", func() {
	Describe("ExtractCredentials", func() {
		It("extracts valid credentials", func() {
			vcapServices := `{"iotf-service":[{"name":"iotf","label":"iotf-service","tags":["internet_of_things","ibm_created"],"plan":"iotf-service-free","credentials":{"iotCredentialsIdentifier":"a2g6k39sl6r5","mqtt_host":"br2ybi.messaging.internetofthings.ibmcloud.com","mqtt_u_port":1883,"mqtt_s_port":8883,"base_uri":"https://internetofthings.ibmcloud.com:443/api/v0001","org":"br2ybi","apiKey":"a-br2ybi-y0tc7vicym","apiToken":"AJIpvsdJ!a__nqR(TK"}}]}`

			creds, err := ExtractCredentials(vcapServices)
			Expect(err).NotTo(HaveOccurred())
			Expect(creds.User).To(Equal("a-br2ybi-y0tc7vicym"))
		})

		It("errors with empty VCAP_SERVICES", func() {
			vcapServices := "{}"

			_, err := ExtractCredentials(vcapServices)
			Expect(err).To(HaveOccurred())
		})

		It("errors with empty string", func() {
			vcapServices := ""

			_, err := ExtractCredentials(vcapServices)
			Expect(err).To(HaveOccurred())
		})

	})
})
