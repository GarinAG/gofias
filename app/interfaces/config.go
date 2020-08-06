package interfaces

type ConfigInterface interface {
	Init() error
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	GetFloat64(key string) float64
}
