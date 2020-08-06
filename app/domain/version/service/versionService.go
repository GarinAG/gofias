package service

import (
	"github.com/GarinAG/gofias/domain/version/entity"
	"github.com/GarinAG/gofias/domain/version/repository"
	"github.com/GarinAG/gofias/interfaces"
	"os"
)

type VersionService struct {
	versionRepo repository.VersionRepositoryInterface
	logger      interfaces.LoggerInterface
}

func NewVersionService(versionRepo repository.VersionRepositoryInterface, logger interfaces.LoggerInterface) *VersionService {
	err := versionRepo.Init()
	if err != nil {
		logger.Fatal(err.Error())
		os.Exit(1)
	}

	return &VersionService{
		versionRepo: versionRepo,
		logger:      logger,
	}
}

func (v *VersionService) GetLastVersionInfo() *entity.Version {
	version, err := v.versionRepo.GetVersion()
	if err != nil {
		v.logger.Fatal(err.Error())
		os.Exit(1)
	}

	return version
}

func (v *VersionService) UpdateVersion(version *entity.Version) {
	err := v.versionRepo.SetVersion(version)
	if err != nil {
		v.logger.Fatal(err.Error())
		os.Exit(1)
	}
}
