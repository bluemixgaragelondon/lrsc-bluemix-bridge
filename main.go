package main

import (
	"github.com/cromega/clogger"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/iotf"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/reporter"
	"os"
)

var logger clogger.Logger
var lrscClient lrscConnection

func main() {
	logger = createLogger()
	iotf.Logger = logger

	logger.Info("================ LRSC <-> IoTF bridge launched  ==================")

	reporters, err := startBridge()
	if err != nil {
		logger.Fatal(err.Error())
	}

	setupHttp(reporters)
	err = startHttp()
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
}

func startBridge() (map[string]*reporter.StatusReporter, error) {
	appReporter := reporter.New()

	// brokerConnection, commands, events, err := setupIotfClient(&appReporter)
	// if err != nil {
	// return nil, err
	// }

	if err := setupLrscClient(); err != nil {
		return nil, err
	}

	reporters := make(map[string]*reporter.StatusReporter)
	reporters["app"] = &appReporter
	reporters["lrsc"] = &lrscClient.StatusReporter
	// reporters["broker"] = &brokerConnection.StatusReporter

	go runConnectionLoop("LRSC client", &lrscClient)
	// go runConnectionLoop("IoTF client", brokerConnection)

	// go func() {
	// for commandMessage := range commands {
	// logger.Debug("Received command message: %v", commandMessage)
	// }
	// }()

	// go func() {
	// for {
	// message := <-lrscClient.inbound
	// event := iotf.Event{Device: message.Deveui, Payload: message.Pdu}
	// events <- event
	// }
	// }()

	return reporters, nil
}

// func setupIotfClient(appReporter *reporter.StatusReporter) (*iotf.BrokerConnection, <-chan iotf.Command, chan<- iotf.Event, error) {
// logger.Info("Starting IoTF connection")
// iotfCreds, err := iotf.ExtractCredentials(os.Getenv("VCAP_SERVICES"))
// if err != nil {
// appReporter.Report("IoTF Credentials", err.Error())
// return nil, nil, nil, err
// }

// commandChannel := make(chan iotf.Command)
// eventChannel := make(chan iotf.Event)

// brokerConnection := iotf.NewBrokerConnection(iotfCreds, eventChannel, commandChannel)

// return brokerConnection, commandChannel, eventChannel, nil
// }

func setupLrscClient() error {
	lrscClient.StatusReporter = reporter.New()
	dialerConfig := dialerConfig{
		host: os.Getenv("LRSC_HOST"),
		port: os.Getenv("LRSC_PORT"),
		cert: os.Getenv("LRSC_CLIENT_CERT"),
		key:  os.Getenv("LRSC_CLIENT_KEY"),
	}
	dialer, err := createTlsDialer(dialerConfig, &lrscClient.StatusReporter)
	if err != nil {
		logger.Error("failed to create dialer: %v", err)
		lrscClient.Report("CONNECTION", err.Error())
		return err
	}

	lrscClient.dialer = dialer
	lrscClient.err = make(chan error)
	lrscClient.inbound = make(chan lrscMessage, 100)

	return nil
}
