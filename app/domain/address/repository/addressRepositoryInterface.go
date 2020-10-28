package repository

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"sync"
	"time"
)

// Интерфейс репозитория адресов
type AddressRepositoryInterface interface {
	// Инициализация таблицы в БД
	Init() error
	// Очистка таблицы в БД
	Clear() error
	// Найти адрес по названию
	GetByFormalName(term string) (*entity.AddressObject, error)
	// Найти город по названию
	GetCityByFormalName(term string) (*entity.AddressObject, error)
	// Найти адрес по GUID
	GetByGuid(guid string) (*entity.AddressObject, error)
	// Найти адреса по GUID
	GetAddressByGuidList(guids []string) ([]*entity.AddressObject, error)
	// Получить список всех городов
	GetCities() ([]*entity.AddressObject, error)
	// Найти города по подстроке
	GetCitiesByTerm(term string, size int64, from int64) ([]*entity.AddressObject, error)
	// Найти адрес по подстроке
	GetAddressByTerm(term string, size int64, from int64) ([]*entity.AddressObject, error)
	// Найти адрес по почтовому индексу
	GetAddressByPostal(term string, size int64, from int64) ([]*entity.AddressObject, error)
	// Найти ближайший город по координатам
	GetNearestCity(lon float64, lat float64) (*entity.AddressObject, error)
	// Найти ближайший адрес по координатам
	GetNearestAddress(lon float64, lat float64, term string) (*entity.AddressObject, error)
	// Обновить коллекцию адресов
	InsertUpdateCollection(wg *sync.WaitGroup, channel <-chan interface{}, count chan<- int, isFull bool)
	// Получить название таблицы в БД
	GetIndexName() string
	// Подсчитать количество адресов в БД по фильтру
	CountAllData(query interface{}) (int64, error)
	// Индексация таблицы
	Index(isFull bool, start time.Time, guids []string, indexChan chan<- entity.IndexObject) error
}
