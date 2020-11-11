package repository

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"sync"
	"time"
)

// Интерфейс функции получения адресов по GUID
type GetIndexObjects func(guids []string) map[string]entity.IndexObject

// Интерфейс репозитория домов
type HouseRepositoryInterface interface {
	// Инициализация таблицы в БД
	Init() error
	// Очистка таблицы в БД
	Clear() error
	// Найти дом по GUID
	GetByGuid(guid string) (*entity.HouseObject, error)
	// Найти дома по GUID адреса
	GetByAddressGuid(guid string) ([]*entity.HouseObject, error)
	// Получить GUID последних обновленных домов
	GetLastUpdatedGuids(start time.Time) ([]string, error)
	// Найти дома по подстроке
	GetAddressByTerm(term string, size int64, from int64, filter ...entity.FilterObject) ([]*entity.HouseObject, error)
	// Обновить коллекцию домов
	InsertUpdateCollection(wg *sync.WaitGroup, channel <-chan interface{}, count chan<- int, isFull bool)
	// Получить название таблицы в БД
	GetIndexName() string
	// Подсчитать количество домов в БД по фильтру
	CountAllData(query interface{}) (int64, error)
	// Индексация таблицы домов
	Index(start time.Time, indexChan <-chan entity.IndexObject, GetIndexObjects GetIndexObjects) error
}
