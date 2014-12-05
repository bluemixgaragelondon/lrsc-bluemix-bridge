package main

import (
	"encoding/json"
	"fmt"
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	"os"
)

var iotfClient *MQTT.MqttClient

var logger, logErr = syslog.Dial("tcp", "logs2.papertrailapp.com:45777", syslog.LOG_SYSLOG|syslog.LOG_INFO, "bridge")

func main() {
	if logErr != nil {
		fmt.Println(logErr)
	}

	iotfCreds := extractIotfCreds(os.Getenv("VCAP_SERVICES"))
	iotfClient = connectToIotf(iotfCreds)

	cert, err := ioutil.ReadFile(os.Getenv("CLIENT_CERT"))
	if err != nil {
		logger.Err(err.Error())
		panic(err)
	}
	key, err := ioutil.ReadFile(os.Getenv("CLIENT_KEY"))
	if err != nil {
		logger.Err(err.Error())
		panic(err)
	}

	dialer, err := CreateTlsDialer("dev.lrsc.ch", "55055", cert, key)
	if err != nil {
		logger.Err(err.Error())
		panic(err)
	}

	lrscConn := &LrscConnection{dialer: dialer}
	messages := make(chan string)
	lrscConn.StartListening(messages)

	go func() {
		for {
			logger.Info("message received: " + <-messages)
		}
	}()

	http.HandleFunc("/", hello)
	http.HandleFunc("/env", env)
	http.HandleFunc("/testpublish", testPublish)

	err = http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		logger.Err(err.Error())
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
		logger.Err(err.Error())
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
		logger.Err(fmt.Sprintf("%v (probably missing configuration)", err))
	}

	iotfBindings := servicesJson["iotf-service"].([]interface{})
	if err != nil {
		logger.Err(err.Error())
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
