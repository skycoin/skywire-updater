package logger

import "github.com/sirupsen/logrus"

// Logger is a custom logger based on logrus logger
type Logger struct {
	*logrus.Entry
}

// NewLogger returns a logger that appends a log field with "service {given name}"
func NewLogger(name string) *Logger {
	logger := &Logger{
		Entry: logrus.WithField("service", name),
	}

	return logger
}
