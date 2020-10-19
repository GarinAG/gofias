package repository

import "github.com/GarinAG/gofias/domain/fiasApi/entity"

// Интерфейс репозитория БД ФИАС
type FiasApiRepositoryInterface interface {
	// Получить все версии БД ФИАС
	GetAllDownloadFileInfo() ([]entity.DownloadFileInfo, error)
	// Получить последнюю версию БД ФИАС
	GetLastDownloadFileInfo() (entity.DownloadFileInfo, error)
}
