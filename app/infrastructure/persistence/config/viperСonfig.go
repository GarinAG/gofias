package config

import (
	"github.com/GarinAG/gofias/interfaces"
	"github.com/spf13/viper"
	"strings"
)

// Объект конфигурации
type ViperConfig struct {
	ConfigPath string
	ConfigType string `default:"env"`
	baseConfig interfaces.BaseConfig
}

// Инициализация конфигурации
func (config *ViperConfig) Init() error {
	viper.AddConfigPath(config.ConfigPath)
	viper.SetConfigType(config.ConfigType)
	var err error
	// Загрузка настроек окружения
	if config.isEnv() {
		viper.SetConfigFile(".env")
		viper.AutomaticEnv()
		_ = viper.ReadInConfig()
	} else {
		err = viper.ReadInConfig()
	}
	config.initConfig()

	return err
}

// Инициализация базовых настроек приложения
func (config *ViperConfig) initConfig() {
	config.baseConfig = interfaces.BaseConfig{
		ProjectPrefix: config.GetString("project.prefix"),
		Elastic: interfaces.ElasticConfig{
			Scheme:   config.GetString("elastic.scheme", "http"),
			Host:     config.GetString("elastic.host", "localhost"),
			Sniff:    config.GetBool("elastic.sniff"),
			Gzip:     config.GetBool("elastic.gzip"),
			User:     config.GetString("elastic.username"),
			Password: config.GetString("elastic.password"),
		},
		BatchSize:         config.GetInt("batch.size", 5000),
		DirectoryFilePath: config.GetString("directory.filePath", "/tmp/fias/"),
		ProcessPrint:      config.GetBool("process.print"),
		FiasApiUrl:        config.GetString("fiasApi.url", "https://fias.nalog.ru/WebServices/Public/"),
		LoggerConsole: interfaces.LoggerConfig{
			Enable: config.GetBool("logger.console.enable"),
			Level:  config.GetString("logger.console.level", "debug"),
			Json:   config.GetBool("logger.console.json"),
			Path:   "",
		},
		LoggerFile: interfaces.LoggerConfig{
			Enable: config.GetBool("logger.file.enable"),
			Level:  config.GetString("logger.file.level", "info"),
			Json:   config.GetBool("logger.file.json"),
			Path:   config.GetString("logger.file.path", "./logs/"),
		},
		Grpc: interfaces.GrpcConfig{
			Network: config.GetString("grpc.network", "tcp"),
			Address: config.GetString("grpc.address", "localhost"),
			Port:    config.GetString("grpc.port", "50051"),
			Gateway: interfaces.GrpcGatewayConfig{
				Enable:  config.GetBool("grpc.gateway.enable"),
				Address: config.GetString("grpc.gateway.address", "localhost"),
				Port:    config.GetString("grpc.gateway.port", "8081"),
			},
			SaveRequest:  config.GetBool("grpc.saveRequest"),
			SaveResponse: config.GetBool("grpc.saveResponse"),
		},
		Workers: interfaces.WorkersConfig{
			Houses:    config.GetInt("workers.houses", 8),
			Addresses: config.GetInt("workers.addresses", 4),
		},
		Osm: interfaces.OsmConfig{
			Url: config.GetString("osm.url", "http://download.geofabrik.de/russia-latest.osm.pbf"),
		},
	}
}

// Получить строку
func (config *ViperConfig) GetString(key string, def ...string) string {
	value := viper.GetString(config.prepareKey(key))
	if value == "" && len(def) > 0 {
		value = def[0]
	}

	return value
}

// Получить булево значение
func (config *ViperConfig) GetBool(key string) bool {
	return viper.GetBool(config.prepareKey(key))
}

// Получить целое число
func (config *ViperConfig) GetInt(key string, def ...int) int {
	value := viper.GetInt(config.prepareKey(key))
	if value == 0 && len(def) > 0 {
		value = def[0]
	}

	return value
}

// Получить дробное число
func (config *ViperConfig) GetFloat64(key string, def ...float64) float64 {
	value := viper.GetFloat64(config.prepareKey(key))
	if value == 0 && len(def) > 0 {
		value = def[0]
	}

	return value
}

// форматирование ключей конфига
func (config *ViperConfig) prepareKey(key string) string {
	if config.isEnv() {
		key = strings.ReplaceAll(key, ".", "_")
		key = strings.ToUpper(key)
	}

	return key
}

// Проверка подключения .env файлов
func (config *ViperConfig) isEnv() bool {
	return config.ConfigType == "env"
}

// Получить конфиги приложения
func (config *ViperConfig) GetConfig() interfaces.BaseConfig {
	return config.baseConfig
}
