package main

import (
	. "github.com/onsi/gomega"
	"testing"
)

func Test_StatusReporter_ReportsStatus(test *testing.T) {
	RegisterTestingT(test)

	reporter := StatusReporter{status: make(map[string]string)}
	reporter.Report("summary", "ok")
	Expect(reporter.status["summary"]).To(Equal("ok"))
}

func Test_StatusReporter_StatusIsJson(test *testing.T) {
	RegisterTestingT(test)

	reporter := StatusReporter{status: make(map[string]string)}
	reporter.Report("summary", "ok")
	Expect(reporter.Status()).To(Equal(`{"summary":"ok"}`))
}
