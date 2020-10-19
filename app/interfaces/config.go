package interfaces

// Интерфейс конфигурации
type ConfigInterface interface {
	// Инициализация конфигурации
	Init() error
	// Получить строку
	GetString(key string) string
	// Получить булево значение
	GetBool(key string) bool
	// Получить целое число
	GetInt(key string) int
	// Получить дробное число
	GetFloat64(key string) float64
}
