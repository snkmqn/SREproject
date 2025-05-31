package logger

import (
	"log"
)

type StdLogger struct{}

func (l *StdLogger) Info(msg string) {
	log.Printf("[INFO] %s", msg)
}

func (l *StdLogger) Infof(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

func (l *StdLogger) Error(msg string) {
	log.Printf("[ERROR] %s", msg)
}

func (l *StdLogger) Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}
