package main

import (
	"encoding/json"
	"fmt"
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"net/http"
	"os"
)

func main() {
	fmt.Printf("%v", extractIotfCreds(os.Getenv("VCAP_SERVICES")))
	http.HandleFunc("/", hello)
	http.HandleFunc("/env", env)
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
	_ = MQTT.NewClientOptions().AddBroker("tcp://iot.eclipse.org:1883")
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

func extractIotfCreds(services string) map[string]string {
	servicesJson := make(map[string]interface{})
	err := json.Unmarshal([]byte(services), &servicesJson)
	if err != nil {
		panic(err)
	}

	iotfBindings := servicesJson["iotf-service"].([]interface{})
	if err != nil {
		panic(err)
	}
	iotf := iotfBindings[0].(map[string]interface{})

	iotfCreds := iotf["credentials"].(map[string]interface{})
	conf := make(map[string]string)
	conf["user"] = iotfCreds["apiKey"].(string)
	conf["password"] = iotfCreds["apiToken"].(string)
	conf["uri"] = fmt.Sprintf("tcp://%v:%v", iotfCreds["mqtt_host"], iotfCreds["mqtt_u_port"])
	return conf
}
