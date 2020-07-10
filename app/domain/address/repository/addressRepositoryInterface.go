package repository

import "github.com/GarinAG/gofias/domain/address/entity"

type AddressRepositoryInterface interface {
	Init() error
	Clear() error
	GetByFormalName(term string) (*entity.AddressObject, error)
	GetCityByFormalName(term string) (*entity.AddressObject, error)
	GetByGuid(guid string) (*entity.AddressObject, error)
	GetCities() ([]*entity.AddressObject, error)
	GetCitiesByTerm(term string, count int64) ([]*entity.AddressObject, error)
	InsertUpdateCollection(collection []interface{}, isFull bool) error
	Flush(full bool, params ...interface{}) error
}
