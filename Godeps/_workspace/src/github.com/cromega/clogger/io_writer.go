package clogger

import (
	"fmt"
	"io"
	"log"
	"time"
)

type writerLogger struct {
	logger *log.Logger
	baseIo io.Closer
	level  int
}

func CreateIoWriter(target io.WriteCloser) Logger {
	logger := log.New(target, "", 0)
	return &writerLogger{logger: logger, baseIo: target, level: Warning}
}

func (l *writerLogger) log(level int, message string) {
	if l.level < level {
		return
	}

	var severity string

	switch level {
	case 1:
		severity = "FATAL"
	case 2:
		severity = "ERROR"
	case 3:
		severity = "WARNING"
	case 4:
		severity = "INFO"
	case 5:
		severity = "DEBUG"
	}

	now := time.Now()
	timestamp := now.Format("Mon 01 15:04:05")

	l.logger.Printf("%v: %v: %v", timestamp, severity, message)
}

func (l *writerLogger) Debug(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log(Debug, message)
}

func (l *writerLogger) Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log(Info, message)
}

func (l *writerLogger) Warning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log(Warning, message)
}

func (l *writerLogger) Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log(Error, message)
}

func (l *writerLogger) Fatal(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log(Fatal, message)
}

func (l *writerLogger) SetLevel(level int) {
	l.level = level
}

func (l *writerLogger) Close() {
	l.baseIo.Close()
}
