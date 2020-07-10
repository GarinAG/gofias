package logger

import (
	"github.com/GarinAG/gofias/interfaces"
	"github.com/sirupsen/logrus"
	"os"
)

type LogrusLogger struct {
	logger *logrus.Logger
}

func InitLogrusLogger(configInterface interfaces.ConfigInterface) interfaces.LoggerInterface {
	logger := logrus.Logger{}
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)

	return &LogrusLogger{
		logger: &logger,
	}
}

func (l *LogrusLogger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(l.prepareArgs(msg, args...)...)
}

func (l *LogrusLogger) Info(msg string, args ...interface{}) {
	l.logger.Info(l.prepareArgs(msg, args...)...)
}

func (l *LogrusLogger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(l.prepareArgs(msg, args...)...)
}

func (l *LogrusLogger) Error(msg string, args ...interface{}) {
	l.logger.Error(l.prepareArgs(msg, args...)...)
}

func (l *LogrusLogger) Panic(msg string, args ...interface{}) {
	l.logger.Panic(l.prepareArgs(msg, args...)...)
}

func (l *LogrusLogger) Fatal(msg string, args ...interface{}) {
	l.logger.Fatal(l.prepareArgs(msg, args...)...)
}

func (l *LogrusLogger) prepareArgs(msg string, args ...interface{}) []interface{} {
	return append([]interface{}{msg}, args)
}
