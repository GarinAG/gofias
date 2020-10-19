package service

import (
	"github.com/GarinAG/gofias/domain/fiasApi/entity"
	"github.com/GarinAG/gofias/domain/fiasApi/repository"
	"github.com/GarinAG/gofias/interfaces"
	"os"
)

// Сервис работы с ФИАС
type FiasApiService struct {
	logger      interfaces.LoggerInterface            // Логгер
	fiasApiRepo repository.FiasApiRepositoryInterface // Репозиторий БД ФИАС
}

// Инициализация сервиса
func NewFiasApiService(fiasApiRepo repository.FiasApiRepositoryInterface, logger interfaces.LoggerInterface) *FiasApiService {
	return &FiasApiService{
		logger:      logger,
		fiasApiRepo: fiasApiRepo,
	}
}

// Получить все версии БД ФИАС
func (f *FiasApiService) GetAllDownloadFileInfo() []entity.DownloadFileInfo {
	res, err := f.fiasApiRepo.GetAllDownloadFileInfo()
	f.checkFatalError(err)

	return res
}

// Получить последнюю версию БД ФИАС
func (f *FiasApiService) GetLastDownloadFileInfo() entity.DownloadFileInfo {
	res, err := f.fiasApiRepo.GetLastDownloadFileInfo()
	f.checkFatalError(err)

	return res
}

// Проверяет наличие ошибки и логирует ее
func (f *FiasApiService) checkFatalError(err error) {
	if err != nil {
		f.logger.Fatal(err.Error())
		os.Exit(1)
	}
}
