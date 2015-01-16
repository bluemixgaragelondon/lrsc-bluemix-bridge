package clogger

const (
	Off = iota
	Fatal
	Error
	Warning
	Info
	Debug
)

type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
	SetLevel(level int)
	Close()
}
