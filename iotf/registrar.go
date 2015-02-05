package iotf

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type deviceRegistrar interface {
	registerDevice(deviceId string) error
	deviceRegistered(deviceId string) bool
}

type iotfHttpRegistrar struct {
	credentials       *Credentials
	devicesRegistered map[string]struct{}
	deviceType        string
}

func newIotfHttpRegistrar(credentials *Credentials, deviceType string) *iotfHttpRegistrar {
	devicesRegistered := make(map[string]struct{})
	return &iotfHttpRegistrar{credentials: credentials, devicesRegistered: devicesRegistered, deviceType: deviceType}
}

func (self *iotfHttpRegistrar) registerDevice(deviceId string) error {
	if self.deviceRegistered(deviceId) {
		return nil
	}

	logger.Debug("Registering new device %v", deviceId)
	url := fmt.Sprintf("%v/organizations/%v/devices", self.credentials.BaseUri, self.credentials.Org)
	body := strings.NewReader(fmt.Sprintf(`{"id":"%v", "type": "%v"}`, deviceId, self.deviceType))
	request, _ := http.NewRequest("POST", url, body)

	request.SetBasicAuth(self.credentials.User, self.credentials.Password)

	request.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	response, _ := client.Do(request)
	defer response.Body.Close()

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

	self.devicesRegistered[deviceId] = struct{}{}
	return nil
}

func (self *iotfHttpRegistrar) deviceRegistered(deviceId string) bool {
	_, deviceRegistered := self.devicesRegistered[deviceId]
	return deviceRegistered
}
