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
			registrar = newRegistrar(credentials)

		})

		AfterEach(func() {
			server.Close()
		})

		Context("the device is registered successfully", func() {

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
						ghttp.VerifyJSON(`{"id":"123456789", "type": "LRSC"}`),
						ghttp.RespondWith(http.StatusCreated, nil, nil),
					),
				)
				err := registrar.registerDevice("123456789", "LRSC")
				Expect(err).To(Succeed())
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("the device already exists", func() {
			It("succeeds", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", registrationPath),
						ghttp.VerifyJSON(`{"id":"123456789", "type": "LRSC"}`),
						ghttp.RespondWith(http.StatusConflict, nil, nil),
					),
				)
				err := registrar.registerDevice("123456789", "LRSC")
				Expect(err).To(Succeed())
			})
		})

		Context("the device is not created", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", registrationPath),
						ghttp.RespondWith(http.StatusForbidden, nil, nil),
					),
				)
				err := registrar.registerDevice("", "")
				Expect(err).To(HaveOccurred())
			})
		})

	})
})
