package entity

// Объект разбора OSM-файла
type Node struct {
	Type       string
	Name       string
	Lat        float64
	Lon        float64
	PostalCode string
}
