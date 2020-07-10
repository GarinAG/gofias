package service

import (
	"github.com/GarinAG/gofias/domain/version/entity"
	"github.com/GarinAG/gofias/domain/version/repository"
	"github.com/GarinAG/gofias/interfaces"
)

type VersionService struct {
	versionRepo repository.VersionRepositoryInterface
	logger      interfaces.LoggerInterface
}

func NewVersionService(versionRepo repository.VersionRepositoryInterface, logger interfaces.LoggerInterface) *VersionService {
	return &VersionService{
		versionRepo: versionRepo,
		logger:      logger,
	}
}

func (v *VersionService) GetLastVersionInfo() *entity.Version {
	version, err := v.versionRepo.GetVersion()
	if err != nil {
		v.logger.Error(err.Error())
	}

	return version
}

func (v *VersionService) UpdateVersion(version *entity.Version) bool {
	err := v.versionRepo.SetVersion(version)
	if err != nil {
		v.logger.Error(err.Error())
		return false
	}

	return true
}
