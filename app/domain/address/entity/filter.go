package entity

// Объект фильтра
type FilterObject struct {
	Level      NumberFilter
	ParentGuid StringFilter
	KladrId    StringFilter
}

// Строковый фильтр
type StringFilter struct {
	Values []string
}

// Числовой фильтр
type NumberFilter struct {
	Values []float32
	Min    float32
	Max    float32
}
