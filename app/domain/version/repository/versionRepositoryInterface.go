package repository

import "github.com/GarinAG/gofias/domain/version/entity"

type VersionRepositoryInterface interface {
	Init() error
	Clear() error
	GetVersion() (*entity.Version, error)
	SetVersion(version *entity.Version) error
}
