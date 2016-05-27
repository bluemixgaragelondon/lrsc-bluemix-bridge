package mqtt

import (
	paho "github.com/eclipse/paho.mqtt.golang"
)

type pahoClient struct {
	client *paho.Client
}

func NewPahoClient(options ClientOptions) Client {
	pahoOptions := paho.NewClientOptions()

	pahoOptions.AddBroker(options.Broker)
	pahoOptions.SetClientID(options.ClientId)
	pahoOptions.SetUsername(options.Username)
	pahoOptions.SetPassword(options.Password)

	pahoOptions.SetConnectionLostHandler(func(client paho.Client, err error) {
		options.OnConnectionLost(err)
	})

	newClient := paho.NewClient(pahoOptions)

	return &pahoClient{client: &newClient}
}

func (self *pahoClient) Start() error {
	client := *self.client
	token := client.Connect()
	token.Wait()
	return token.Error()
}

func (self *pahoClient) PublishMessage(topic string, message []byte) {
	client := *self.client
	client.Publish(topic, 1, false, message)
}

func (self *pahoClient) StartSubscription(topic string, callback func(message Message)) error {
  client := *self.client
	token := client.Subscribe(topic, 0, func(client paho.Client, message paho.Message) {
		callback(message)
	})
	
	token.Wait()
	return token.Error()
}
