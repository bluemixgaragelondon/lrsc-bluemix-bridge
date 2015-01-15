package iotf

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"net/http"
)

var _ = Describe("Registrar", func() {
	Describe("registerDevice", func() {
		var (
			server           *ghttp.Server
			registrar        deviceRegistrar
			registrationPath string
		)

		BeforeEach(func() {
			server = ghttp.NewServer()
			credentials := Credentials{
				User:     "testuser",
				Password: "testpass",
				Org:      "testorg",
				BaseUri:  server.URL()}

			registrationPath = "/organizations/testorg/devices"
			registrar = &iotfHttpRegistrar{&credentials}

		})

		AfterEach(func() {
			server.Close()
		})

		It("sends credentials", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", registrationPath),
					ghttp.VerifyBasicAuth("testuser", "testpass"),
					ghttp.RespondWith(http.StatusCreated, nil, nil),
				),
			)
			err := registrar.registerDevice("", "")
			Expect(err).To(Succeed())
		})

		It("POSTs the device information", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", registrationPath),
					ghttp.VerifyJSON(`{"id":"123456789", "type": "TEST"}`),
					ghttp.RespondWith(http.StatusCreated, nil, nil),
				),
			)
			err := registrar.registerDevice("TEST", "123456789")
			Expect(err).To(Succeed())
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})

		Context("the device is not in IoTF", func() {
			It("succeeds", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", registrationPath),
						ghttp.VerifyJSON(`{"id":"123456789", "type": "TEST"}`),
						ghttp.RespondWith(http.StatusCreated, nil, nil),
					),
				)
				err := registrar.registerDevice("TEST", "123456789")
				Expect(err).To(Succeed())
			})
		})

		Context("the device already exists in IoTF", func() {
			It("succeeds", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", registrationPath),
						ghttp.VerifyJSON(`{"id":"123456789", "type": "TEST"}`),
						ghttp.RespondWith(http.StatusConflict, nil, nil),
					),
				)
				err := registrar.registerDevice("TEST", "123456789")
				Expect(err).To(Succeed())
			})
		})

		Context("the IoTF service is broken", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", registrationPath),
						ghttp.RespondWith(http.StatusInternalServerError, nil, nil),
					),
				)
				err := registrar.registerDevice("", "")
				Expect(err).To(HaveOccurred())
			})
		})

	})
})
