package repository

import "github.com/GarinAG/gofias/domain/version/entity"

// Интефейс репозитория версий
type VersionRepositoryInterface interface {
	// Инициализация таблицы в БД
	Init() error
	// Очистка таблицы в БД
	Clear() error
	// Получить текущую версию БД ФИАС
	GetVersion() (*entity.Version, error)
	// Сохранить версию
	SetVersion(version *entity.Version) error
}
