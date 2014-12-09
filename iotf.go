package main

import (
	"encoding/json"
	"fmt"
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"log"
	"math/rand"
	"os"
	"time"
)

func connectToIotf(iotfCreds map[string]string) *MQTT.MqttClient {
	clientOpts := MQTT.NewClientOptions()
	clientOpts.AddBroker(iotfCreds["uri"])
	clientOpts.SetClientId(fmt.Sprintf("a:%v:$v", iotfCreds["org"], generateClientIdSuffix()))
	clientOpts.SetUsername(iotfCreds["user"])
	clientOpts.SetPassword(iotfCreds["password"])

	clientOpts.SetOnConnectionLost(func(client *MQTT.MqttClient, err error) {
		logger.Error("IoTF connection lost handler called: " + err.Error())
	})

	MQTT.WARN = log.New(os.Stdout, "", 0)
	MQTT.ERROR = log.New(os.Stdout, "", 0)
	MQTT.DEBUG = log.New(os.Stdout, "", 0)

	client := MQTT.NewClient(clientOpts)
	_, err := client.Start()
	if err != nil {
		logger.Error(err.Error())
	}

	return client
}

func generateClientIdSuffix() string {
	rand.Seed(time.Now().UTC().UnixNano())
	suffix := rand.Intn(1000)
	return string(suffix)
}

func extractIotfCreds(services string) map[string]string {
	servicesJson := make(map[string]interface{})
	err := json.Unmarshal([]byte(services), &servicesJson)
	if err != nil {
		logger.Error(fmt.Sprintf("%v (probably missing configuration)", err))
	}

	iotfBindings := servicesJson["iotf-service"].([]interface{})
	if err != nil {
		logger.Error(err.Error())
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
