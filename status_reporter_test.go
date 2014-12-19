package main

import (
	. "github.com/onsi/gomega"
	"testing"
)

func Test_StatusReporter_ReportsStatus(test *testing.T) {
	RegisterTestingT(test)

	reporter := statusReporter{stats: make(map[string]string)}
	reporter.report("summary", "ok")
	Expect(reporter.stats["summary"]).To(Equal("ok"))
}

func Test_StatusReporter_StatusIsJson(test *testing.T) {
	RegisterTestingT(test)

	reporter := statusReporter{stats: make(map[string]string)}
	reporter.report("summary", "ok")
	Expect(reporter.status()).To(Equal(`{"summary":"ok"}`))
}
