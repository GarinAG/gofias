package service

import (
	"github.com/GarinAG/gofias/domain/fiasApi/entity"
	"github.com/GarinAG/gofias/domain/fiasApi/repository"
	"github.com/GarinAG/gofias/interfaces"
	"os"
)

type FiasApiService struct {
	logger      interfaces.LoggerInterface
	fiasApiRepo repository.FiasApiRepositoryInterface
}

func NewFiasApiService(fiasApiRepo repository.FiasApiRepositoryInterface, logger interfaces.LoggerInterface) *FiasApiService {
	return &FiasApiService{
		logger:      logger,
		fiasApiRepo: fiasApiRepo,
	}
}

func (f *FiasApiService) GetAllDownloadFileInfo() []entity.DownloadFileInfo {
	res, err := f.fiasApiRepo.GetAllDownloadFileInfo()
	if err != nil {
		f.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("GetAllDownloadFileInfo error")
		os.Exit(1)
	}

	return res
}

func (f *FiasApiService) GetLastDownloadFileInfo() entity.DownloadFileInfo {
	res, err := f.fiasApiRepo.GetLastDownloadFileInfo()
	if err != nil {
		f.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("GetLastDownloadFileInfo error")
		os.Exit(1)
	}

	return res
}
