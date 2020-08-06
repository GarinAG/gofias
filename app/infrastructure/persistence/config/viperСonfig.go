package config

import (
	"github.com/spf13/viper"
	"strings"
)

type ViperConfig struct {
	ConfigPath string
	ConfigType string `default:"env"`
}

func (config *ViperConfig) Init() error {
	viper.AddConfigPath(config.ConfigPath)
	viper.SetConfigType(config.ConfigType)
	if config.isEnv() {
		viper.SetConfigFile(".env")
		viper.AutomaticEnv()
		_ = viper.ReadInConfig()

		return nil
	} else {
		return viper.ReadInConfig()
	}
}

func (config *ViperConfig) GetString(key string) string {
	return viper.GetString(config.prepareKey(key))
}

func (config *ViperConfig) GetBool(key string) bool {
	return viper.GetBool(config.prepareKey(key))
}

func (config *ViperConfig) GetInt(key string) int {
	return viper.GetInt(config.prepareKey(key))
}

func (config *ViperConfig) GetFloat64(key string) float64 {
	return viper.GetFloat64(config.prepareKey(key))
}

func (config *ViperConfig) prepareKey(key string) string {
	if config.isEnv() {
		key = strings.ReplaceAll(key, ".", "_")
		key = strings.ToUpper(key)
	}

	return key
}

func (config *ViperConfig) isEnv() bool {
	return config.ConfigType == "env"
}
