package main

import (
	"encoding/json"
)

type (
	lrscMessageMode int

	lrscMessageType int
)

const (
	messageModeUnconfirmed lrscMessageMode = 0
	messageModeConfirmed   lrscMessageMode = 2

	messageTypeHandshake  lrscMessageType = 0
	messageTypeUpstream   lrscMessageType = 6
	messageTypeDownstream lrscMessageType = 7
)

type lrscMessage struct {
	Type             lrscMessageType `json:"msgtag"`
	DeviceGuid       string          `json:"deveui"`
	Payload          string          `json:"pdu"`
	UniqueSequenceNo uint64          `json:"seqno"`
	Mode             lrscMessageMode `json:"mode"`
	Timeout          uint            `json:"timeout"`
	Port             uint            `json:"port"`
}

func parseLrscMessage(s string) (lrscMessage, error) {
	var message lrscMessage

	err := json.Unmarshal([]byte(s), &message)

	return message, err
}
