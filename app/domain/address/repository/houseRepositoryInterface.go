package repository

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"time"
)

type HouseRepositoryInterface interface {
	Init() error
	Clear() error
	GetByAddressGuid(guid string) ([]*entity.HouseObject, error)
	GetLastUpdatedGuids(start time.Time) ([]string, error)
	GetAddressByTerm(term string, size int64, from int64) ([]*entity.HouseObject, error)
	InsertUpdateCollection(channel <-chan interface{}, done <-chan bool, count chan<- int, isFull bool)
	GetIndexName() string
	CountAllData(query interface{}) (int64, error)
	Index(indexChan <-chan entity.IndexObject) error
}
