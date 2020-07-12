package entity

type Version struct {
	ID               string
	FiasVersion      int
	UpdateDate       string
	RecUpdateAddress int
	RecUpdateHouses  int
}

func (v Version) TableName() string {
	return "fias_version"
}
