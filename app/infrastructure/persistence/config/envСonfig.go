package config

import (
	"os"
	"strconv"
	"strings"
)

type EnvConfig struct{}

func (config *EnvConfig) Init() error {
	return nil
}

func (config *EnvConfig) GetString(key string) string {
	return os.Getenv(config.prepareKey(key))
}

func (config *EnvConfig) GetBool(key string) bool {
	res := os.Getenv(config.prepareKey(key))
	return res == "true" || res == "1"
}

func (config *EnvConfig) GetInt(key string) int {
	res, _ := strconv.Atoi(os.Getenv(config.prepareKey(key)))

	return res
}

func (config *EnvConfig) GetFloat64(key string) float64 {
	res, _ := strconv.ParseFloat(os.Getenv(config.prepareKey(key)), 64)

	return res
}

func (config *EnvConfig) prepareKey(key string) string {
	key = strings.ReplaceAll(key, ".", "_")
	key = strings.ToUpper(key)

	return key
}
