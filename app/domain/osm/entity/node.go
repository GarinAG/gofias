package entity

import "github.com/paulmach/osm"

// Объект разбора OSM-файла
type Node struct {
	Type         string
	Name         string
	HouseAddress string
	Lat          float64
	Lon          float64
	PostalCode   string
	Node         *osm.Node
}
