package dto

type JsonVersionDto struct {
	ID               int    `json:"version_id"`
	FiasVersion      string `json:"fias_version"`
	UpdateDate       string `json:"update_date"`
	RecUpdateAddress int    `json:"rec_upd_address"`
	RecUpdateHouses  int    `json:"rec_upd_houses"`
}
