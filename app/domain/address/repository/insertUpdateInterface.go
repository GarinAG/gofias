package repository

import "sync"

// Интерфейс обновления данных в БД
type InsertUpdateInterface interface {
	// Обновить коллекцию
	InsertUpdateCollection(wg *sync.WaitGroup, channel <-chan interface{}, count chan<- int, isFull bool)
}
