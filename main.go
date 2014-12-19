package main

import (
	"os"
)

var logger = createLogger()
var lrscClient lrscConnection
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
	if err := setupIotfClient(); err != nil {
		return err
	}

	if err := setupLrscClient(); err != nil {
		return err
	}

	go runConnectionLoop("LRSC client", &lrscClient)
	go runConnectionLoop("IoTF client", &iotfClient)

	go func() {
		for {
			message := <-lrscClient.inbound
			iotfClient.publish(message.Deveui, message.Pdu)
		}
	}()

	return nil
}

func setupIotfClient() error {
	iotfClient.stats = make(map[string]string)

	logger.Info("Starting IoTF connection")
	iotfCreds, err := extractIotfCreds(os.Getenv("VCAP_SERVICES"))
	if err != nil {
		iotfClient.report("CONNECTION", err.Error())
		return err
	}
	iotfClient.initialise(iotfCreds, "LRSC")

	return nil
}

func setupLrscClient() error {
	lrscClient.stats = make(map[string]string)
	dialerConfig := dialerConfig{
		host: os.Getenv("LRSC_HOST"),
		port: os.Getenv("LRSC_PORT"),
		cert: os.Getenv("LRSC_CLIENT_CERT"),
		key:  os.Getenv("LRSC_CLIENT_KEY"),
	}
	dialer, err := createTlsDialer(dialerConfig, &lrscClient.statusReporter)
	if err != nil {
		logger.Error("failed to create dialer: %v", err)
		lrscClient.report("CONNECTION", err.Error())
		return err
	}

	lrscClient.dialer = dialer
	lrscClient.err = make(chan error)
	lrscClient.inbound = make(chan lrscMessage, 100)

	return nil
}
