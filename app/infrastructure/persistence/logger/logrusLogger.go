package logger

import (
	"github.com/GarinAG/gofias/interfaces"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
)

type logrusLogger struct {
	logger *logrus.Logger
}

type logrusLogEntry struct {
	entry *logrus.Entry
}

func getFormatter(isJSON bool) logrus.Formatter {
	if isJSON {
		return &logrus.JSONFormatter{}
	}

	return &logrus.TextFormatter{
		FullTimestamp:          true,
		DisableLevelTruncation: true,
	}
}

func NewLogrusLogger(config interfaces.LoggerConfiguration) (interfaces.LoggerInterface, error) {
	logLevel := config.ConsoleLevel
	if logLevel == "" {
		logLevel = config.FileLevel
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}
	stdOutHandler := os.Stdout
	fileHandler := &lumberjack.Logger{
		Filename: config.FileLocation,
		MaxSize:  100,
		Compress: true,
		MaxAge:   28,
	}
	lLogger := &logrus.Logger{
		Out:       stdOutHandler,
		Formatter: getFormatter(config.ConsoleJSONFormat),
		Hooks:     make(logrus.LevelHooks),
		Level:     level,
	}

	if config.EnableConsole && config.EnableFile {
		lLogger.SetOutput(io.MultiWriter(stdOutHandler, fileHandler))
	} else {
		if config.EnableFile {
			lLogger.SetOutput(fileHandler)
			lLogger.SetFormatter(getFormatter(config.FileJSONFormat))
		}
	}

	return &logrusLogger{
		logger: lLogger,
	}, nil
}

func (l *logrusLogger) Debug(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *logrusLogger) Info(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *logrusLogger) Warn(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *logrusLogger) Error(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *logrusLogger) Fatal(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *logrusLogger) Panic(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *logrusLogger) WithFields(fields interfaces.LoggerFields) interfaces.LoggerInterface {
	return &logrusLogEntry{
		entry: l.logger.WithFields(convertToLogrusFields(fields)),
	}
}

func (l *logrusLogEntry) Debug(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

func (l *logrusLogEntry) Info(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l *logrusLogEntry) Warn(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

func (l *logrusLogEntry) Error(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

func (l *logrusLogEntry) Fatal(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

func (l *logrusLogEntry) Panic(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

func (l *logrusLogEntry) WithFields(fields interfaces.LoggerFields) interfaces.LoggerInterface {
	return &logrusLogEntry{
		entry: l.entry.WithFields(convertToLogrusFields(fields)),
	}
}

func convertToLogrusFields(fields interfaces.LoggerFields) logrus.Fields {
	logrusFields := logrus.Fields{}
	for index, val := range fields {
		logrusFields[index] = val
	}
	return logrusFields
}
