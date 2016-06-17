package iotf

import (
	"fmt"
	"github.com/bluemixgaragelondon/lrsc-bluemix-bridge/mqtt"
)

type clientFactory interface {
	newClient(clientId string) mqtt.Client
}

type mqttClientFactory struct {
	credentials           Credentials
	connectionLostHandler func(err error)
}

func (f *mqttClientFactory) newClient(clientId string) mqtt.Client {
	clientOptions := mqtt.ClientOptions{
		Broker:           fmt.Sprintf("tcps://%v:%v", f.credentials.MqttHost, f.credentials.MqttSecurePort),
		ClientId:         fmt.Sprintf("a:%v:%v", f.credentials.Org, clientId),
		Username:         f.credentials.User,
		Password:         f.credentials.Password,
		OnConnectionLost: f.connectionLostHandler,
	}

	return mqtt.NewPahoClient(clientOptions)
}
