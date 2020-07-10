package config

import (
	"github.com/spf13/viper"
)

type YamlConfig struct {
	ConfigPath string
}

func (config *YamlConfig) Init() error {
	viper.AddConfigPath(config.ConfigPath)
	viper.SetConfigName("config")

	return viper.ReadInConfig()
}

func (config *YamlConfig) GetString(key string) string {
	return viper.GetString(key)
}

func (config *YamlConfig) GetBool(key string) bool {
	return viper.GetBool(key)
}

func (config *YamlConfig) GetInt(key string) int {
	return viper.GetInt(key)
}

func (config *YamlConfig) GetFloat64(key string) float64 {
	return viper.GetFloat64(key)
}
