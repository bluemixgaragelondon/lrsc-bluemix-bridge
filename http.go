package main

import (
	"fmt"
	"github.com/bluemixgaragelondon/lrsc-bluemix-bridge/reporter"
	"net/http"
	"os"
	"runtime"
)

func setupHttp(reporters map[string]reporter.StatusReporter) {
	http.Handle("/", http.FileServer(http.Dir("public")))
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

	http.HandleFunc("/stack", func(res http.ResponseWriter, req *http.Request) {
		data := make([]byte, 100000)
		all := true
		length := runtime.Stack(data, all)
		fmt.Fprintf(res, "%v", string(data[:length]))
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
