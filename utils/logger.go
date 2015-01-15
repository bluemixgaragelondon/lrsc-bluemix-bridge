package utils

import (
	"fmt"
	"github.com/cromega/clogger"
	"net/url"
	"os"
)

func CreateLogger() clogger.Logger {
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
