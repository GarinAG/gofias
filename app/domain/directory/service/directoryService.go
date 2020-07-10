package service

import (
	"github.com/GarinAG/gofias/domain/directory/entity"
	"github.com/GarinAG/gofias/interfaces"
)

type DirectoryService struct {
	logger          interfaces.LoggerInterface
	downloadService *DownloadService
}

func NewDirectoryService(logger interfaces.LoggerInterface, config interfaces.ConfigInterface) *DirectoryService {
	return &DirectoryService{
		logger:          logger,
		downloadService: NewDownloadService(logger, config),
	}
}

func (d *DirectoryService) DownloadAndExtractFile(url string, parts ...string) *[]entity.File {
	file, err := d.downloadService.DownloadFile(url)
	if err != nil {
		d.logger.Fatal(err.Error())
	}
	extractedFiles, err := d.downloadService.Unzip(file, parts...)
	if err != nil {
		d.logger.Fatal(err.Error())
	}

	return &extractedFiles
}
