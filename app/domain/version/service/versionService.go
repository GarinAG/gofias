package service

import (
	"github.com/GarinAG/gofias/domain/version/entity"
	"github.com/GarinAG/gofias/domain/version/repository"
	"github.com/GarinAG/gofias/interfaces"
	"os"
)

// Сервис управления версиями
type VersionService struct {
	versionRepo repository.VersionRepositoryInterface // Репозиторий версий БД
	logger      interfaces.LoggerInterface            // Логгер
}

// Инициализация сервиса
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

// Получить последнюю скачанную версию
func (v *VersionService) GetLastVersionInfo() *entity.Version {
	version, err := v.versionRepo.GetVersion()
	v.checkFatalError(err)

	return version
}

// Обновить версию
func (v *VersionService) UpdateVersion(version *entity.Version) {
	err := v.versionRepo.SetVersion(version)
	v.checkFatalError(err)
}

// Проверяет наличие ошибки и логирует ее
func (v *VersionService) checkFatalError(err error) {
	if err != nil {
		v.logger.Fatal(err.Error())
		os.Exit(1)
	}
}
