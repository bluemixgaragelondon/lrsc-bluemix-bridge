package main

import (
	"encoding/json"
)

type StatusReporter struct {
	status map[string]string
}

func (self *StatusReporter) Report(key, value string) {
	self.status[key] = value
}

func (self StatusReporter) Status() string {
	summary, _ := json.Marshal(self.status)
	return string(summary)
}
