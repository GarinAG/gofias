package logger

import (
	"github.com/GarinAG/gofias/interfaces"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Объект-обёртка над логгером Logrus
type logrusLogger struct {
	logger *logrus.Logger
}

// Объект-обёртка над логгером Logrus для сохранения логов с доп. полями
type logrusLogEntry struct {
	entry *logrus.Entry
}

// Установить правила форматирование логов
func getFormatter(isJSON bool) logrus.Formatter {
	if isJSON {
		return &logrus.JSONFormatter{}
	}

	return &logrus.TextFormatter{
		FullTimestamp:          true,
		DisableLevelTruncation: true,
	}
}

// Инициализация логгера
func NewLogrusLogger(config interfaces.LoggerConfiguration) interfaces.LoggerInterface {
	logLevel := config.ConsoleLevel
	if logLevel == "" {
		logLevel = config.FileLevel
	}

	// Устанавливает уровень логирования
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		panic(err)
	}
	stdOutHandler := os.Stdout
	// Создает обработчик сохранения логов в файл
	filePath := filepath.Dir(config.FileLocation) + "/" + config.FileLocationPrefix + "/log-" + time.Now().Format("2006-01-02") + ".log"
	fileHandler := &lumberjack.Logger{
		Filename: filePath,
		MaxSize:  100,
		Compress: true,
		MaxAge:   28,
	}
	// Создает обработчик вывода логов в консоль
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
		} else {
			lLogger.SetOutput(stdOutHandler)
		}
	}

	return &logrusLogger{
		logger: lLogger,
	}
}

// Вывести отладку
func (l *logrusLogger) Debug(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

// Вывести информацию
func (l *logrusLogger) Info(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

// Вывести предупреждение
func (l *logrusLogger) Warn(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

// Вывести ошибку
func (l *logrusLogger) Error(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

// Вывести критическую ошибку
func (l *logrusLogger) Fatal(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

// Вывести критическую ошибку
func (l *logrusLogger) Panic(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

// Вывести текст
func (l *logrusLogger) Printf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

// Вывести дополнительные данные
func (l *logrusLogger) WithFields(fields interfaces.LoggerFields) interfaces.LoggerInterface {
	return &logrusLogEntry{
		entry: l.logger.WithFields(convertToLogrusFields(fields)),
	}
}

// Вывести отладку
func (l *logrusLogEntry) Debug(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

// Вывести информацию
func (l *logrusLogEntry) Info(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

// Вывести предупреждение
func (l *logrusLogEntry) Warn(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

// Вывести ошибку
func (l *logrusLogEntry) Error(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

// Вывести критическую ошибку
func (l *logrusLogEntry) Fatal(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

// Вывести критическую ошибку
func (l *logrusLogEntry) Panic(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

// Вывести текст
func (l *logrusLogEntry) Printf(format string, args ...interface{}) {
	l.entry.Printf(format, args...)
}

// Сохранить доп. поля
func (l *logrusLogEntry) WithFields(fields interfaces.LoggerFields) interfaces.LoggerInterface {
	return &logrusLogEntry{
		entry: l.entry.WithFields(convertToLogrusFields(fields)),
	}
}

// Конвертирует интерфейс в поля данных логгера
func convertToLogrusFields(fields interfaces.LoggerFields) logrus.Fields {
	logrusFields := logrus.Fields{}
	for index, val := range fields {
		logrusFields[index] = val
	}
	return logrusFields
}
