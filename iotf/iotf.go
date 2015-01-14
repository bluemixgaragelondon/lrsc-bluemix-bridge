package iotf

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Event struct {
	Device, Payload string
}

type Command struct {
	Device, Payload string
}

type Credentials struct {
	User             string `json:"apiKey"`
	Password         string `json:"apiToken"`
	Org              string
	BaseUri          string `json:"base_uri"`
	MqttHost         string `json:"mqtt_host"`
	MqttSecurePort   int    `json:"mqtt_s_port"`
	MqttUnsecurePort int    `json:"mqtt_u_port"`
}

func ExtractCredentials(services string) (*Credentials, error) {
	data := struct {
		Services []struct {
			Credentials Credentials
		} `json:"iotf-service"`
	}{}

	err := json.Unmarshal([]byte(services), &data)
	if err != nil {
		return nil, fmt.Errorf("Could not parse services JSON: %v", err)
	}

	if len(data.Services) == 0 {
		return nil, errors.New("Could not find any iotf-service instance bound")
	}

	return &data.Services[0].Credentials, nil
}
