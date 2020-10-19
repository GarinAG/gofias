package repository

import (
	"github.com/GarinAG/gofias/domain/address/entity"
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
	// Получить список всех городов
	GetCities() ([]*entity.AddressObject, error)
	// Найти города по подстроке
	GetCitiesByTerm(term string, size int64, from int64) ([]*entity.AddressObject, error)
	// Найти адрес по подстроке
	GetAddressByTerm(term string, size int64, from int64) ([]*entity.AddressObject, error)
	// Найти адрес по почтовому индексу
	GetAddressByPostal(term string, size int64, from int64) ([]*entity.AddressObject, error)
	// Обновить коллекцию адресов
	InsertUpdateCollection(channel <-chan interface{}, done <-chan bool, count chan<- int, isFull bool)
	// Получить название таблицы в БД
	GetIndexName() string
	// Подсчитать количество адресов в БД по фильтру
	CountAllData(query interface{}) (int64, error)
	// Индексация таблицы
	Index(isFull bool, start time.Time, guids []string, indexChan chan<- entity.IndexObject) error
}
