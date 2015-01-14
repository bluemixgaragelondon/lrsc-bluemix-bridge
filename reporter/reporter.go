package reporter

import (
	"encoding/json"
)

type StatusReporter struct {
	stats map[string]string
}

func (self *StatusReporter) Report(key, value string) {
	self.stats[key] = value
}

func (self StatusReporter) Summary() string {
	summary, _ := json.Marshal(self.stats)
	return string(summary)
}

func New() StatusReporter {
	return StatusReporter{stats: make(map[string]string)}
}
