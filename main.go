package main

import (
	"fmt"
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/cromega/clogger"
	"io/ioutil"
	"net/url"
	"os"
)

var iotfTopic = "iot-2/type/Dummy/id/lrsc-client-test-sensor-1/evt/TEST/fmt/json"
var iotfClient *MQTT.MqttClient

var logger = createLogger()

func main() {
	cert, err := ioutil.ReadFile(os.Getenv("CLIENT_CERT"))
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
	key, err := ioutil.ReadFile(os.Getenv("CLIENT_KEY"))
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}

	dialer, err := CreateTlsDialer("dev.lrsc.ch", "55055", cert, key)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}

	lrscConn := &LrscConnection{dialer: dialer}
	messages := make(chan lrscMessage)
	lrscConn.StartListening(messages)

	iotfCreds := extractIotfCreds(os.Getenv("VCAP_SERVICES"))
	iotfClient = connectToIotf(iotfCreds)

	go func() {
		for {
			message := <-messages
			logger.Info("Received message %v from device %v", message.Pdu, message.Deveui)
			mqttMessage := MQTT.NewMessage([]byte(message.Pdu))
			iotfClient.PublishMessage(iotfTopic, mqttMessage)
		}
	}()

	setupHttp()
	startHttp()
}

func createLogger() clogger.Logger {
	logTarget := os.Getenv("REMOTE_LOG_URL")
	var logger clogger.Logger

	if logTarget == "" {
		logger = clogger.CreateIoWriter(os.Stdout)
	} else {
		uri, err := url.Parse(logTarget)
		if err != nil {
			fmt.Printf("Error: failed to parse remote log target (%v)", err)
		}

		app := "lrsc-bridge-" + os.Getenv("LRSC_ENV")

		logger, err = clogger.CreateSyslog(uri.Scheme, uri.Host, app)
		if err != nil {
			fmt.Printf("Error: failed to initialise remote logger (%v)", err)
		}
	}

	return logger
}
