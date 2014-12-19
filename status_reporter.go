package main

import (
	"encoding/json"
)

type statusReporter struct {
	stats map[string]string
}

func (self *statusReporter) report(key, value string) {
	self.stats[key] = value
}

func (self statusReporter) status() string {
	summary, _ := json.Marshal(self.stats)
	return string(summary)
}
