package main

import (
	"encoding/json"
)

type lrscMessageMode int

const (
	messageModeUnconfirmed lrscMessageMode = 0
	messageModeConfirmed   lrscMessageMode = 2
)

type lrscMessage struct {
	DeviceGuid       string          `json:"deveui"`
	Payload          string          `json:"pdu"`
	UniqueSequenceNo int             `json:"seqno"`
	Mode             lrscMessageMode `json:"mode"`
	Timeout          int             `json:"timeout"`
	Port             int             `json:"port"`
}

func parseLrscMessage(data string) (lrscMessage, error) {
	var message lrscMessage
	err := json.Unmarshal([]byte(data), &message)
	return message, err
}
