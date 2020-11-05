package entity

type FilterObject struct {
	Level      NumberFilter
	ParentGuid StringFilter
	KladrId    StringFilter
}

type StringFilter struct {
	Values []string
}

type NumberFilter struct {
	Values []float32
	Min    float32
	Max    float32
}
