package reporter

import (
	"encoding/json"
)

type StatusReporter interface {
	Report(key, value string)
	Summary() string
}

type BridgeReporter struct {
	stats map[string]string
}

func (self *BridgeReporter) Report(key, value string) {
	self.stats[key] = value
}

func (self *BridgeReporter) Summary() string {
	summary, _ := json.Marshal(self.stats)
	return string(summary)
}

func New() StatusReporter {
	return &BridgeReporter{stats: make(map[string]string)}
}
