package logger

import (
	"github.com/GarinAG/gofias/interfaces"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"time"
)

// Объект-обёртка над логгером Zap
type zapLogger struct {
	sugaredLogger *zap.SugaredLogger
}

// Установить формат даты
func zapEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05Z07:00"))
}

// Установить правила форматирование логов
func getEncoder(isJSON bool) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapEncoder
	encoderConfig.TimeKey = "time"
	if isJSON {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// Конвертирует уровень логов в формат Zap
func getZapLevel(level string) zapcore.Level {
	switch level {
	case interfaces.Info:
		return zapcore.InfoLevel
	case interfaces.Warn:
		return zapcore.WarnLevel
	case interfaces.Debug:
		return zapcore.DebugLevel
	case interfaces.Error:
		return zapcore.ErrorLevel
	case interfaces.Fatal:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Инициализация логгера
func NewZapLogger(config interfaces.LoggerConfiguration) interfaces.LoggerInterface {
	var cores []zapcore.Core

	// Создает обработчик вывода логов в консоль
	if config.EnableConsole {
		level := getZapLevel(config.ConsoleLevel)
		writer := zapcore.Lock(os.Stdout)
		core := zapcore.NewCore(getEncoder(config.ConsoleJSONFormat), writer, level)
		cores = append(cores, core)
	}

	// Создает обработчик сохранения логов в файл
	if config.EnableFile {
		filePath := filepath.Dir(config.FileLocation) + "/" + config.FileLocationPrefix + "/log-" + time.Now().Format("2006-01-02") + ".log"
		level := getZapLevel(config.FileLevel)
		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename: filePath,
			MaxSize:  100,
			Compress: true,
			MaxAge:   28,
		})
		core := zapcore.NewCore(getEncoder(config.FileJSONFormat), writer, level)
		cores = append(cores, core)
	}

	combinedCore := zapcore.NewTee(cores...)

	logger := zap.New(combinedCore,
		zap.AddCallerSkip(2),
		zap.AddCaller(),
	).Sugar()

	return &zapLogger{
		sugaredLogger: logger,
	}
}

// Вывести отладку
func (l *zapLogger) Debug(format string, args ...interface{}) {
	l.sugaredLogger.Debugf(format, args...)
}

// Вывести информацию
func (l *zapLogger) Info(format string, args ...interface{}) {
	l.sugaredLogger.Infof(format, args...)
}

// Вывести предупреждение
func (l *zapLogger) Warn(format string, args ...interface{}) {
	l.sugaredLogger.Warnf(format, args...)
}

// Вывести ошибку
func (l *zapLogger) Error(format string, args ...interface{}) {
	l.sugaredLogger.Errorf(format, args...)
}

// Вывести критическую ошибку
func (l *zapLogger) Fatal(format string, args ...interface{}) {
	l.sugaredLogger.Fatalf(format, args...)
}

// Вывести критическую ошибку
func (l *zapLogger) Panic(format string, args ...interface{}) {
	l.sugaredLogger.Fatalf(format, args...)
}

// Вывести текст
func (l *zapLogger) Printf(format string, args ...interface{}) {
	l.sugaredLogger.Infof(format, args...)
}

// Вывести дополнительные данные
func (l *zapLogger) WithFields(fields interfaces.LoggerFields) interfaces.LoggerInterface {
	var f = make([]interface{}, 0)
	for k, v := range fields {
		f = append(f, k)
		f = append(f, v)
	}
	newLogger := l.sugaredLogger.With(f...)
	return &zapLogger{newLogger}
}
