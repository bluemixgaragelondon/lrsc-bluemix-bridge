package iotf

type Event struct {
	Device, Payload string
}

type Command struct {
	Device, Payload string
}

type iotfCredentials struct {
	User             string `json:"apiKey"`
	Password         string `json:"apiToken"`
	Org              string
	BaseUri          string `json:"base_uri"`
	MqttHost         string `json:"mqtt_host"`
	MqttSecurePort   int    `json:"mqtt_s_port"`
	MqttUnsecurePort int    `json:"mqtt_u_port"`
}
