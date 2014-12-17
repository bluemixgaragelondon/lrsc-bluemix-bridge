package main

import (
	"fmt"
	"github.com/cromega/clogger"
	"net/url"
	"os"
)

var logger = createLogger()
var lrscClient LrscConnection
var iotfClient iotfConnection

func main() {
	logger.Info("================ LRSC <-> IoTF bridge launched  ==================")

	err := startBridge()
	if err != nil {
		logger.Fatal(err.Error())
	}

	setupHttp()
	err = startHttp()
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
}

func startBridge() error {
	setupReporting(&iotfClient.StatusReporter)
	setupReporting(&lrscClient.StatusReporter)

	logger.Info("Starting IoTF connection")
	iotfCreds, err := extractIotfCreds(os.Getenv("VCAP_SERVICES"))
	if err != nil {
		iotfClient.Report("CONNECTION", err.Error())
		return err
	}
	iotfClient.Initialise(iotfCreds, "LRSC")
	err = iotfClient.Connect()
	if err != nil {
		return err
	}
	logger.Info("Established IoTF connection")

	dialerConfig := dialerConfig{
		host: os.Getenv("LRSC_HOST"),
		port: os.Getenv("LRSC_PORT"),
		cert: os.Getenv("LRSC_CLIENT_CERT"),
		key:  os.Getenv("LRSC_CLIENT_KEY"),
	}
	dialer, err := CreateTlsDialer(dialerConfig, &lrscClient.StatusReporter)
	if err != nil {
		lrscClient.Report("CONNECTION", err.Error())
		return err
	}

	lrscClient.dialer = dialer
	messages := make(chan lrscMessage)

	logger.Info("Starting LRSC connection")
	lrscClient.StartListening(messages)

	go listenForMessages(messages)
	return nil
}

func setupReporting(reporter *StatusReporter) {
	reporter.status = make(map[string]string)
}

func listenForMessages(messages chan lrscMessage) {
	for {
		message := <-messages
		logger.Info("Received message %v from device %v", message.Pdu, message.Deveui)
		iotfClient.Publish(message.Deveui, message.toJson())
	}
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
