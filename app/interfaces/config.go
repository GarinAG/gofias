package interfaces

// Интерфейс конфигурации
type ConfigInterface interface {
	// Инициализация конфигурации
	Init() error
	// Получить строку
	GetString(key string, def ...string) string
	// Получить булево значение
	GetBool(key string) bool
	// Получить целое число
	GetInt(key string, def ...int) int
	// Получить дробное число
	GetFloat64(key string, def ...float64) float64
	// Получить конфиги приложения
	GetConfig() BaseConfig
}

// Конфиги эластика
type ElasticConfig struct {
	Scheme   string // Протокол соединения
	Host     string // Хост
	Sniff    bool   // Запрашивать статус эластика
	Gzip     bool   // Сжимать данные
	User     string // Пользователь
	Password string // Пароль пользователя
}

// Конфиги логгера
type LoggerConfig struct {
	Enable bool   // Активность логгера
	Level  string // Уровень логов
	Json   bool   // Форматировать в JSON
	Path   string // Путь к файлам логов
}

// Конфиги Grpc-сервера
type GrpcConfig struct {
	Network      string // Протокол соединения
	Address      string // Хост для запуска сервера
	Port         string // Порт
	SaveRequest  bool
	SaveResponse bool
	Gateway      GrpcGatewayConfig // Конфиги RestApi-сервера
}

// Конфиги RestApi-сервера
type GrpcGatewayConfig struct {
	Enable  bool   // Активность сервера
	Address string // Хост для запуска сервера
	Port    string // Порт
}

// Конфиги пула обработчиков
type WorkersConfig struct {
	Houses    int // Количество обработчиков для домов
	Addresses int // Количество обработчиков для адресов
}

// Конфиги OSM
type OsmConfig struct {
	Url string // Путь до файла с данными
}

// Базовые конфиги приложения
type BaseConfig struct {
	ProjectPrefix     string        // Префикс проекта для хранения в БД
	Elastic           ElasticConfig // Конфиги эластика
	BatchSize         int           // Размер пачки для обновления
	DirectoryFilePath string        // Путь сохранения файлов импорта
	MaxTries          int           // Максимальное количество попыток скачивания файлов/запросов к API
	ProcessPrint      bool          // Разрешить вывод прогресса в консоль
	FiasApiUrl        string        // Путь до FIAS Api сервиса
	LoggerConsole     LoggerConfig  // Конфиги консольного логгера
	LoggerFile        LoggerConfig  // Конфиги файлового логгера
	Grpc              GrpcConfig    // Конфиги grpc-сервера
	Workers           WorkersConfig // Конфиги RestApi-сервера
	Osm               OsmConfig     // Конфиги OSM (гео-данные)
}
