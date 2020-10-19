package repository

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"time"
)

// Интерфейс репозитория домов
type HouseRepositoryInterface interface {
	// Инициализация таблицы в БД
	Init() error
	// Очистка таблицы в БД
	Clear() error
	// Найти дома по GUID адреса
	GetByAddressGuid(guid string) ([]*entity.HouseObject, error)
	// Получить GUID последних обновленных домов
	GetLastUpdatedGuids(start time.Time) ([]string, error)
	// Найти дома по подстроке
	GetAddressByTerm(term string, size int64, from int64) ([]*entity.HouseObject, error)
	// Обновить коллекцию домов
	InsertUpdateCollection(channel <-chan interface{}, done <-chan bool, count chan<- int, isFull bool)
	// Получить название таблицы в БД
	GetIndexName() string
	// Подсчитать количество домов в БД по фильтру
	CountAllData(query interface{}) (int64, error)
	// Индексация таблицы домов
	Index(indexChan <-chan entity.IndexObject) error
}
