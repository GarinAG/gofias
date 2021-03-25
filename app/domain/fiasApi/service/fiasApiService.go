package service

import (
	"github.com/GarinAG/gofias/domain/fiasApi/entity"
	"github.com/GarinAG/gofias/domain/fiasApi/repository"
	"github.com/GarinAG/gofias/interfaces"
	"os"
	"time"
)

// Сервис работы с ФИАС
type FiasApiService struct {
	logger      interfaces.LoggerInterface            // Логгер
	fiasApiRepo repository.FiasApiRepositoryInterface // Репозиторий БД ФИАС
	config      interfaces.ConfigInterface
}

// Инициализация сервиса
func NewFiasApiService(fiasApiRepo repository.FiasApiRepositoryInterface, logger interfaces.LoggerInterface, config interfaces.ConfigInterface) *FiasApiService {
	return &FiasApiService{
		logger:      logger,
		fiasApiRepo: fiasApiRepo,
		config:      config,
	}
}

// Получить все версии БД ФИАС
func (f *FiasApiService) GetAllDownloadFileInfo() []entity.DownloadFileInfo {
	// Максимальное количество попыток запросов
	maxTries := f.config.GetConfig().MaxTries
	var res []entity.DownloadFileInfo
	var err error

	for i := 1; i <= maxTries; i++ {
		res, err = f.fiasApiRepo.GetAllDownloadFileInfo()
		if err == nil {
			break
		}
		// Повторный запрос
		if i < maxTries {
			f.logger.Error(err.Error())
			f.logger.WithFields(interfaces.LoggerFields{"attempt": i + 1}).Info("Trying to re-download info")
			time.Sleep(5 * time.Second)
		}
	}
	f.checkFatalError(err)

	return res
}

// Получить последнюю версию БД ФИАС
func (f *FiasApiService) GetLastDownloadFileInfo() entity.DownloadFileInfo {
	// Максимальное количество попыток запросов
	maxTries := f.config.GetConfig().MaxTries
	var res entity.DownloadFileInfo
	var err error

	for i := 1; i <= maxTries; i++ {
		res, err = f.fiasApiRepo.GetLastDownloadFileInfo()
		if err == nil {
			break
		}
		// Повторный запрос
		if i < maxTries {
			f.logger.Error(err.Error())
			f.logger.WithFields(interfaces.LoggerFields{"attempt": i + 1}).Info("Trying to re-download info")
			time.Sleep(5 * time.Second)
		}
	}
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
