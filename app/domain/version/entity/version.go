package entity

// Объект версии
type Version struct {
	ID               int
	FiasVersion      string
	UpdateDate       string
	RecUpdateAddress int
	RecUpdateHouses  int
}

// Получить название таблицы в БД
func (v Version) TableName() string {
	return "fias_version"
}
