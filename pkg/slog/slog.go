package slog

import (
	"github.com/sirupsen/logrus"
)

func SetLevel(level logrus.Level) {
	logrus.SetLevel(level)
}

func Logger() *logrus.Logger {
	return logrus.StandardLogger()
}

func Copy() *logrus.Logger {
	currentLogger := Logger()
	logger := logrus.New()
	logger.SetFormatter(currentLogger.Formatter)
	logger.SetLevel(currentLogger.Level)
	return logger
}

func SetFormatter(isDev bool) {
	var formatter logrus.Formatter = &logrus.JSONFormatter{}
	if isDev {
		formatter = &logrus.TextFormatter{}
	}
	logrus.SetFormatter(formatter)
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Error(args ...interface{}) {
	logrus.Error(args...)
}

func Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}

func Fatal(args ...interface{}) {
	logrus.Fatal(args...)
}

func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func Info(args ...interface{}) {
	logrus.Info(args...)
}
