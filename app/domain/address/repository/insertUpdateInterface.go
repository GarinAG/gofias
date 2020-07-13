package repository

type InsertUpdateInterface interface {
	InsertUpdateCollection(channel <-chan interface{}, done <-chan bool, count chan<- int)
}
