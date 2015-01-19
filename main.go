package main

import (
	"fmt"
	"github.com/cromega/clogger"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/iotf"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/reporter"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/utils"
	"os"
	"time"
)

var logger clogger.Logger
var lrscClient lrscConnection

func init() {
	logger = utils.CreateLogger()
}

func main() {

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

	commands := make(chan iotf.Command)
	events := make(chan iotf.Event)

	iotfManager, err := iotf.NewIoTFManager(os.Getenv("VCAP_SERVICES"), commands, events)
	if err != nil {
		appReporter.Report("IoTF Manager:", err.Error())
		return nil, err
	}

	if err := setupLrscClient(); err != nil {
		return nil, err
	}

	reporters := make(map[string]*reporter.StatusReporter)
	reporters["app"] = &appReporter
	reporters["lrsc"] = &lrscClient.StatusReporter
	reporters["iotf"] = iotfManager.StatusReporter()

	go runConnectionLoop("LRSC client", &lrscClient)
	go runConnectionLoop("IoTF client", iotfManager)

	go func() {
		for commandMessage := range commands {
			logger.Debug("Received command message: %v", commandMessage)
		}
	}()

	go func() {
		for {
			message := <-lrscClient.inbound
			event := iotf.Event{Device: message.Deveui, Payload: message.Pdu}
			events <- event
		}
	}()

	return reporters, nil
}

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

type connection interface {
	Connect() error
	Error() <-chan error
	Loop()
}

func retryWithBackoff(name string, body func() error) {
	for timeout := time.Second; ; timeout *= 2 {
		logger.Debug("starting " + name)
		if timeout > time.Minute*5 {
			timeout = time.Minute * 5
		}

		err := body()
		if err == nil {
			return
		} else {
			time.Sleep(timeout)
			logger.Debug("connection failed, retrying " + name)
		}
	}
}

func runConnectionLoop(name string, conn connection) {
	logger.Debug("starting connection loop " + name)
	for {
		retryWithBackoff(name, conn.Connect)
		go conn.Loop()
		err := <-conn.Error()
		logger.Error(fmt.Sprintf("%v went tits up: %v", name, err))
	}
}
