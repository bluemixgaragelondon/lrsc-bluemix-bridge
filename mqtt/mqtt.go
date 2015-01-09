package mqtt

type Client interface {
	Start() error
	PublishMessage(topic string, message []byte)
	StartSubscription(topic string, callback func(message Message)) error
}

type Message interface {
	Topic() string
	Payload() []byte
}

type ClientOptions struct {
	Broker, ClientId   string
	Username, Password string
	OnConnectionLost   func(error)
}
