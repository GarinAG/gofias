package repository

import "github.com/GarinAG/gofias/domain/fiasApi/entity"

type FiasApiRepositoryInterface interface {
	GetAllDownloadFileInfo() ([]entity.DownloadFileInfo, error)
	GetLastDownloadFileInfo() (entity.DownloadFileInfo, error)
}
