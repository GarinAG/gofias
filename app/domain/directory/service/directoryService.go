package service

import (
	"github.com/GarinAG/gofias/domain/directory/entity"
	"github.com/GarinAG/gofias/interfaces"
	"os"
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
	// Скачивает файл
	file, err := d.downloadService.DownloadFile(url, fileName)
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
