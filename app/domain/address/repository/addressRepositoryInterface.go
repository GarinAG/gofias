package repository

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"time"
)

type AddressRepositoryInterface interface {
	Init() error
	Clear() error
	GetByFormalName(term string) (*entity.AddressObject, error)
	GetCityByFormalName(term string) (*entity.AddressObject, error)
	GetByGuid(guid string) (*entity.AddressObject, error)
	GetCities() ([]*entity.AddressObject, error)
	GetCitiesByTerm(term string, count int64) ([]*entity.AddressObject, error)
	InsertUpdateCollection(channel chan interface{}, done chan bool, count chan int) error
	Index(houseRepos HouseRepositoryInterface, isFull bool, start time.Time) error
	GetIndexName() string
}
