package repository

import (
	"github.com/GarinAG/gofias/domain/address/entity"
)

type HouseRepositoryInterface interface {
	Init() error
	Clear() error
	GetByAddressGuid(guid string) ([]*entity.HouseObject, error)
	InsertUpdateCollection(channel <-chan interface{}, done <-chan bool, count chan<- int, isFull bool)
	GetIndexName() string
	CountAllData(query interface{}) (int64, error)
}
