package clogger

import (
	"fmt"
	"log/syslog"
)

type syslogLogger struct {
	logger *syslog.Writer
	level  int
}

func CreateSyslog(proto, raddr, prefix string) (Logger, error) {
	logger, err := syslog.Dial(proto, raddr, syslog.LOG_CRIT, prefix)
	return &syslogLogger{logger: logger, level: Warning}, err
}

func (l syslogLogger) Debug(format string, args ...interface{}) {
	if l.level < Debug {
		return
	}

	message := fmt.Sprintf(format, args...)
	l.logger.Debug(message)
}

func (l syslogLogger) Info(format string, args ...interface{}) {
	if l.level < Info {
		return
	}

	message := fmt.Sprintf(format, args...)
	l.logger.Info(message)
}

func (l syslogLogger) Warning(format string, args ...interface{}) {
	if l.level < Warning {
		return
	}

	message := fmt.Sprintf(format, args...)
	l.logger.Warning(message)
}

func (l syslogLogger) Error(format string, args ...interface{}) {
	if l.level < Error {
		return
	}

	message := fmt.Sprintf(format, args...)
	l.logger.Err(message)
}

func (l syslogLogger) Fatal(format string, args ...interface{}) {
	if l.level < Fatal {
		return
	}

	message := fmt.Sprintf(format, args...)
	l.logger.Crit(message)
}

func (l *syslogLogger) SetLevel(level int) {
	if level < Fatal {
		level = Fatal
	}

	l.level = level
}

func (l *syslogLogger) Close() {
	l.logger.Close()
}
