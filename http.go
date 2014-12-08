package main

import (
	"fmt"
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"net/http"
	"os"
)

func setupHttp() {
	http.HandleFunc("/", hello)
	http.HandleFunc("/env", env)
	http.HandleFunc("/testpublish", testPublish)
}

func startHttp() {
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
}

func hello(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "hello, world")
}

func env(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain")
	for key, value := range os.Environ() {
		fmt.Fprintf(res, "%v = %v\n", key, value)
	}
}

func testPublish(res http.ResponseWriter, req *http.Request) {
	message := MQTT.NewMessage([]byte(`{"msg": "Hello world"}`))
	iotfClient.PublishMessage(iotfTopic, message)
	fmt.Fprintf(res, "done")
}
