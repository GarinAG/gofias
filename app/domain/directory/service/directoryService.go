package service

import (
	"github.com/GarinAG/gofias/domain/directory/entity"
	"github.com/GarinAG/gofias/interfaces"
	"os"
	"time"
)

// Сервис работы с файлами
type DirectoryService struct {
	logger          interfaces.LoggerInterface // Логгер
	downloadService *DownloadService           // Сервис управления загрузкой файлов
}

// Инициализация сервиса
func NewDirectoryService(downloadService *DownloadService, logger interfaces.LoggerInterface, config interfaces.ConfigInterface) *DirectoryService {
	return &DirectoryService{
		logger:          logger,
		downloadService: downloadService,
	}
}

// Очистка директории
func (d *DirectoryService) ClearDirectory() {
	d.downloadService.ClearDirectory()
}

// Скачать и распаковать файлы
func (d *DirectoryService) DownloadAndExtractFile(url string, fileName string, parts ...string) *[]entity.File {
	// Максимальное количество попыток скачивания файла
	maxTries := d.downloadService.config.GetConfig().MaxTries
	var file *entity.File
	var err error
	for i := 1; i <= maxTries; i++ {
		// Скачивает файл
		file, err = d.downloadService.DownloadFile(url, fileName)
		if file != nil {
			break
		}
		// Повторное скачивание файла при ошибке
		if i < maxTries {
			if err != nil {
				d.logger.Error(err.Error())
			}
			d.logger.WithFields(interfaces.LoggerFields{"file": url, "attempt": i + 1}).Info("Trying to re-download file")
			time.Sleep(5 * time.Second)
		}
	}
	d.checkFatalError(err)
	// Распаковывает файл
	extractedFiles, err := d.downloadService.Unzip(file, parts...)
	d.checkFatalError(err)

	return &extractedFiles
}

// Проверяет наличие ошибки и логирует ее
func (d *DirectoryService) checkFatalError(err error) {
	if err != nil {
		d.logger.Fatal(err.Error())
		os.Exit(1)
	}
}
