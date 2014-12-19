package main

import (
	"fmt"
	"github.com/cromega/clogger"
	"net/url"
	"os"
	"time"
)

type connection interface {
	connect() error
	error() chan error
	loop()
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
	logger.Debug("starting conneciton loop " + name)
	for {
		retryWithBackoff(name, conn.connect)
		go conn.loop()
		err := <-conn.error()
		logger.Error(fmt.Sprintf("%v went tits up: %v", name, err))
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

	logger.SetLevel(clogger.Debug)

	return logger
}
