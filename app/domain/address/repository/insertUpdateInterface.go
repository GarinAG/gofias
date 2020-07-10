package repository

type InsertUpdateInterface interface {
	InsertUpdateCollection(collection []interface{}, isFull bool) error
}
