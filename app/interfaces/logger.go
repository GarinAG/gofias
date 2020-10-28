package interfaces

// Интерфейс набора дополнительных полей
type LoggerFields map[string]interface{}

const (
	// Вывод отладочной информации
	Debug = "debug"
	// Вывод логов по умолчанию
	Info = "info"
	// Вывод сообщений о возможных проблемах
	Warn = "warn"
	// Вывод ошибок
	Error = "error"
	// Вывод критических ошибок. Приложение завершает работу после регистрации сообщения.
	Fatal = "fatal"
)

// Интерфейс логгера
type LoggerInterface interface {
	// Вывести отладку
	Debug(format string, args ...interface{})
	// Вывести информацию
	Info(format string, args ...interface{})
	// Вывести предупреждение
	Warn(format string, args ...interface{})
	// Вывести ошибку
	Error(format string, args ...interface{})
	// Вывести критическую ошибку
	Fatal(format string, args ...interface{})
	// Вывести критическую ошибку
	Panic(format string, args ...interface{})
	// Вывести дополнительные данные
	WithFields(keyValues LoggerFields) LoggerInterface
	// Вывести текст
	Printf(format string, args ...interface{})
}

// Объект конфигурации логгера
type LoggerConfiguration struct {
	EnableConsole      bool   // Разрешить вывод в консоль
	ConsoleJSONFormat  bool   // Формат вывода в консоль
	ConsoleLevel       string // Уровень ошибок для вывода в консоль
	EnableFile         bool   // Разрешить сохранять в файл
	FileJSONFormat     bool   // Формат сохранения в файл
	FileLevel          string // Уровень ошибок для сохранения в файл
	FileLocation       string // Путь до папки с логами
	FileLocationPrefix string // Префикс названия файла логов
}
