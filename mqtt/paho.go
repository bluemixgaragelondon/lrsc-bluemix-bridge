package mqtt

import (
	paho "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

type pahoClient struct {
	client *paho.MqttClient
}

func NewPahoClient(options ClientOptions) Client {
	pahoOptions := paho.NewClientOptions()

	pahoOptions.AddBroker(options.Broker)
	pahoOptions.SetClientId(options.ClientId)
	pahoOptions.SetUsername(options.Username)
	pahoOptions.SetPassword(options.Password)

	pahoOptions.SetOnConnectionLost(func(client *paho.MqttClient, err error) {
		options.OnConnectionLost(err)
	})

	return &pahoClient{client: paho.NewClient(pahoOptions)}
}

func (self *pahoClient) Start() error {
	_, err := self.client.Start()
	return err
}

func (self *pahoClient) PublishMessage(topic string, message []byte) {
	self.client.PublishMessage(topic, paho.NewMessage(message))
}

func (self *pahoClient) StartSubscription(topic string, callback func(message Message)) error {
	pahoCallback := func(client *paho.MqttClient, pahoMessage paho.Message) {
		// the paho.Message struct implements the Message interface
		callback(&pahoMessage)
	}

	topicFilter, _ := paho.NewTopicFilter(topic, byte(paho.QOS_ZERO))

	_, err := self.client.StartSubscription(pahoCallback, topicFilter)
	return err
}
