package main

import (
	"fmt"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/reporter"
	"net/http"
	"os"
)

func setupHttp(reporters map[string]*reporter.StatusReporter) {
	http.Handle("/", http.FileServer(http.Dir("web")))
	http.HandleFunc("/env", env)

	http.HandleFunc("/iotfStatus", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		reporter, present := reporters["iotf"]
		if present {
			fmt.Fprintf(res, "%v", reporter.Summary())
		}
	})

	http.HandleFunc("/lrscStatus", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		reporter, present := reporters["lrsc"]
		if present {
			fmt.Fprintf(res, "%v", reporter.Summary())
		}
	})
}

func startHttp() error {
	return http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}

func env(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain")
	for key, value := range os.Environ() {
		fmt.Fprintf(res, "%v = %v\n", key, value)
	}
}
