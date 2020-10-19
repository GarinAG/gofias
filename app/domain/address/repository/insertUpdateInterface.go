package repository

// Интерфейс обновления данных в БД
type InsertUpdateInterface interface {
	// Обновить коллекцию
	InsertUpdateCollection(channel <-chan interface{}, done <-chan bool, count chan<- int, isFull bool)
}
