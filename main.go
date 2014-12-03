package main

import (
	"encoding/json"
	"fmt"
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"log"
	"net/http"
	"os"
)

func main() {
	iotfCreds := extractIotfCreds(os.Getenv("VCAP_SERVICES"))
	iotfClient := connectToIotf(iotfCreds)

	testIotfConnection(iotfClient)

	http.HandleFunc("/", hello)
	http.HandleFunc("/env", env)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func connectToIotf(iotfCreds map[string]string) *MQTT.MqttClient {
	clientOpts := MQTT.NewClientOptions()
	clientOpts.AddBroker(iotfCreds["uri"])
	clientOpts.SetClientId(fmt.Sprintf("a:%v:lrsc-client", iotfCreds["org"]))
	clientOpts.SetUsername(iotfCreds["user"])
	clientOpts.SetPassword(iotfCreds["password"])
	MQTT.WARN = log.New(os.Stdout, "", 0)
	MQTT.ERROR = log.New(os.Stdout, "", 0)

	client := MQTT.NewClient(clientOpts)
	_, err := client.Start()
	if err != nil {
		panic(err)
	}

	return client
}

func testIotfConnection(client *MQTT.MqttClient) {
	topic := "iot-2/type/Dummy/id/lrsc-client-test-sensor-1/evt/TEST/fmt/json"
	message := MQTT.NewMessage([]byte(`{"msg": "Hello world"}`))
	client.PublishMessage(topic, message)
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
	conf["uri"] = fmt.Sprintf("tls://%v:%v", iotfCreds["mqtt_host"], iotfCreds["mqtt_s_port"])
	conf["org"] = iotfCreds["org"].(string)
	return conf
}
