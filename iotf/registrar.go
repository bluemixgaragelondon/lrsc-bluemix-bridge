package iotf

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type deviceRegistrar struct {
	credentials Credentials
}

func newRegistrar(credentials Credentials) deviceRegistrar {
	return deviceRegistrar{credentials}
}

func (self *deviceRegistrar) registerDevice(deviceId, deviceType string) error {
	url := fmt.Sprintf("%v/organizations/%v/devices", self.credentials.BaseUri, self.credentials.Org)
	body := strings.NewReader(fmt.Sprintf(`{"id":"%v", "type": "%v"}`, deviceId, deviceType))
	request, _ := http.NewRequest("POST", url, body)

	request.SetBasicAuth(self.credentials.User, self.credentials.Password)

	request.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	response, _ := client.Do(request)

	switch response.StatusCode {
	case http.StatusCreated:
		break
	case http.StatusConflict:
		break
	default:
		return errors.New(fmt.Sprintf("Unable to create device, %d", response.StatusCode))
	}

	return nil
}
