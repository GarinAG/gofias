package service

import (
	"github.com/GarinAG/gofias/domain/directory/entity"
	"github.com/GarinAG/gofias/interfaces"
	"os"
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

func (d *DirectoryService) ClearDirectory() error {
	err := d.downloadService.ClearDirectory()
	if err != nil {
		d.logger.Fatal(err.Error())
		os.Exit(1)
	}

	return err
}

func (d *DirectoryService) DownloadAndExtractFile(url string, parts ...string) *[]entity.File {
	file, err := d.downloadService.DownloadFile(url)
	if err != nil {
		d.logger.Fatal(err.Error())
		os.Exit(1)
	}
	extractedFiles, err := d.downloadService.Unzip(file, parts...)
	if err != nil {
		d.logger.Fatal(err.Error())
		os.Exit(1)
	}

	return &extractedFiles
}
