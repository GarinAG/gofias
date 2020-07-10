package repository

type InsertUpdateInterface interface {
	InsertUpdateCollection(collection []interface{}, isFull bool) error
}

type insertCollection func(repo InsertUpdateInterface, collection []interface{}, node interface{}, isFull bool, size int) []interface{}
