package main

import (
	"encoding/json"
)

type lrscMessage struct {
	DeviceId string `json:"deveui"`
	Payload  string `json:"pdu"`
}

func (m *lrscMessage) toJson() string {
	json, err := json.Marshal(m)
	if err != nil {
		logger.Error("lrscMessage JSON marshaling failed: %v", err.Error())
	}
	return string(json)
}

func parseLrscMessage(data string) (lrscMessage, error) {
	var message lrscMessage
	err := json.Unmarshal([]byte(data), &message)
	return message, err
}
