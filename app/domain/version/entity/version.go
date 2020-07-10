package entity

type Version struct {
	ID               string
	FiasVersion      int
	UpdateDate       string
	RecUpdateAddress string
	RecUpdateHouses  string
}

func (v Version) TableName() string {
	return "fias_version"
}
