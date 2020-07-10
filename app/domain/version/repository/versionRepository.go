package repository

import "github.com/GarinAG/gofias/domain/version/entity"

type VersionRepositoryInterface interface {
	GetVersion() (*entity.Version, error)
	SetVersion(version *entity.Version) error
}
