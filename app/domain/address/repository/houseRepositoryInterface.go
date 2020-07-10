package repository

import "github.com/GarinAG/gofias/domain/address/entity"

type HouseRepositoryInterface interface {
	Init() error
	Clear() error
	GetByAddressGuid(term string) (*entity.HouseObject, error)
	InsertUpdateCollection(collection []interface{}, isFull bool) error
	GetIndexName() string
}
