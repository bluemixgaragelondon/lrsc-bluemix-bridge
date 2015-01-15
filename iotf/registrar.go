package iotf

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type deviceRegistrar interface {
	registerDevice(deviceId, deviceType string) error
}

type iotfHttpRegistrar struct {
	credentials Credentials
}

func (self *iotfHttpRegistrar) registerDevice(deviceId, deviceType string) error {
	logger.Debug("Registering new device %v", deviceId)
	url := fmt.Sprintf("%v/organizations/%v/devices", self.credentials.BaseUri, self.credentials.Org)
	body := strings.NewReader(fmt.Sprintf(`{"id":"%v", "type": "%v"}`, deviceId, deviceType))
	request, _ := http.NewRequest("POST", url, body)

	request.SetBasicAuth(self.credentials.User, self.credentials.Password)

	request.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	response, _ := client.Do(request)

	switch response.StatusCode {
	case http.StatusCreated:
		logger.Debug("Device %v was registered", deviceId)
		break
	case http.StatusConflict:
		logger.Warning("Device %v was already registered", deviceId)
		break
	default:
		logger.Error("Unable to register device (http %v)", response.StatusCode)
		return errors.New(fmt.Sprintf("Unable to create device, %d", response.StatusCode))
	}

	return nil
}
