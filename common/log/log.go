package log

import (
	"fmt"
	"log"
	"time"
)

var defaultLogger Logger = &logger{}

type Logger interface {
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

type logger struct {
}

func printf(level, format string, args ...interface{}) {
	t := time.Now().Format("2006-01-02 15:04:05")
	log.Printf(fmt.Sprintf("[%s] %s %s", level, t, format), args...)
}

func (l *logger) Info(format string, args ...interface{}) {
	printf("INFO", format, args...)
}

func (l *logger) Warn(format string, args ...interface{}) {
	printf("WARN", format, args...)
}

func (l *logger) Error(format string, args ...interface{}) {
	printf("EROR", format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

func SetLogger(l Logger) {
	defaultLogger = l
}

func GetLogger() Logger {
	return defaultLogger
}
