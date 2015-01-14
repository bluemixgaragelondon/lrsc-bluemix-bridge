package reporter

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("reporter", func() {
	Describe("Report", func() {
		var (
			reporter StatusReporter
		)

		BeforeEach(func() {
			reporter = StatusReporter{stats: make(map[string]string)}
		})

		It("adds value to the report", func() {
			reporter.Report("summary", "ok")
			Expect(reporter.stats["summary"]).To(Equal("ok"))
		})

		Describe("Summary", func() {
			It("returns a summary of the values in the reporter in JSON format", func() {
				reporter.Report("summary", "ok")
				Expect(reporter.Summary()).To(Equal(`{"summary":"ok"}`))
			})
		})
	})
})
