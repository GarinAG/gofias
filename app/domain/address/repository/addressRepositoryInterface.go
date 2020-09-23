package repository

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"time"
)

type GetHousesByGuid func(guid string) []*entity.HouseObject
type GetLastUpdatedGuids func(start time.Time) []string

type AddressRepositoryInterface interface {
	Init() error
	Clear() error
	GetByFormalName(term string) (*entity.AddressObject, error)
	GetCityByFormalName(term string) (*entity.AddressObject, error)
	GetByGuid(guid string) (*entity.AddressObject, error)
	GetCities() ([]*entity.AddressObject, error)
	GetCitiesByTerm(term string, size int64, from int64) ([]*entity.AddressObject, error)
	GetAddressByTerm(term string, size int64, from int64) ([]*entity.AddressObject, error)
	GetAddressByPostal(term string, size int64, from int64) ([]*entity.AddressObject, error)
	InsertUpdateCollection(channel <-chan interface{}, done <-chan bool, count chan<- int, isFull bool)
	GetIndexName() string
	CountAllData(query interface{}) (int64, error)
	Index(isFull bool, start time.Time, guids []string, indexChan chan<- entity.IndexObject) error
}
