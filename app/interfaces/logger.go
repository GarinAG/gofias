package interfaces

type LoggerFields map[string]interface{}

const (
	//Debug has verbose message
	Debug = "debug"
	//Info is default log level
	Info = "info"
	//Warn is for logging messages about possible issues
	Warn = "warn"
	//Error is for logging errors
	Error = "error"
	//Fatal is for logging fatal messages. The sytem shutsdown after logging the message.
	Fatal = "fatal"
)

type LoggerInterface interface {
	Debug(format string, args ...interface{})

	Info(format string, args ...interface{})

	Warn(format string, args ...interface{})

	Error(format string, args ...interface{})

	Fatal(format string, args ...interface{})

	Panic(format string, args ...interface{})

	WithFields(keyValues LoggerFields) LoggerInterface
}

type LoggerConfiguration struct {
	EnableConsole      bool
	ConsoleJSONFormat  bool
	ConsoleLevel       string
	EnableFile         bool
	FileJSONFormat     bool
	FileLevel          string
	FileLocation       string
	FileLocationPrefix string
}
