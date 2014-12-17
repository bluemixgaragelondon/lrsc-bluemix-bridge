package main

import (
	"fmt"
	"net/http"
	"os"
)

func setupHttp() {
	http.Handle("/", http.FileServer(http.Dir("web")))
	http.HandleFunc("/env", env)
	http.HandleFunc("/testpublish", testPublish)
	http.HandleFunc("/iotfStatus", iotfStatus)
	http.HandleFunc("/lrscStatus", lrscStatus)
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

func testPublish(res http.ResponseWriter, req *http.Request) {
	iotfClient.Publish("lrsc-client-test-sensor-1", `{"msg": "Hello world"}`)
	fmt.Fprintf(res, "done")
}

func iotfStatus(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(res, "%v", iotfClient.Status())
}

func lrscStatus(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(res, "%v", lrscClient.Status())
}
