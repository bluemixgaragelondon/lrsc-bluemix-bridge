package main

import (
	"encoding/json"
	"fmt"
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var iotfClient *MQTT.MqttClient

var logger = log.New(os.Stdout, "", 0)

func main() {

	iotfCreds := extractIotfCreds(os.Getenv("VCAP_SERVICES"))
	iotfClient = connectToIotf(iotfCreds)

	cert, err := ioutil.ReadFile(os.Getenv("CLIENT_CERT"))
	if err != nil {
		logger.Panic(err)
	}
	key, err := ioutil.ReadFile(os.Getenv("CLIENT_KEY"))
	if err != nil {
		logger.Panic(err)
	}
	lrscConn, err := CreateLrscConnection("dev.lrsc.ch", "55055", cert, key)
	if err != nil {
		logger.Panic(err)
	}
	err = lrscConn.Connect()
	if err != nil {
		logger.Panic(err)
	}

	http.HandleFunc("/", hello)
	http.HandleFunc("/env", env)
	http.HandleFunc("/testpublish", testPublish)

	err = http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		logger.Panic(err)
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
		logger.Panic(err)
	}

	return client
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
	topic := "iot-2/type/Dummy/id/lrsc-client-test-sensor-1/evt/TEST/fmt/json"
	message := MQTT.NewMessage([]byte(`{"msg": "Hello world"}`))
	iotfClient.PublishMessage(topic, message)
	fmt.Fprintf(res, "done")
}

func extractIotfCreds(services string) map[string]string {
	servicesJson := make(map[string]interface{})
	err := json.Unmarshal([]byte(services), &servicesJson)
	if err != nil {
		logger.Panic(fmt.Sprintf("%v (probably missing configuration)", err))
	}

	iotfBindings := servicesJson["iotf-service"].([]interface{})
	if err != nil {
		logger.Panic(err)
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
