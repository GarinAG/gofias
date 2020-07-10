package dto

type JsonVersionDto struct {
	ID               string `json:"version_id"`
	FiasVersion      int    `json:"fias_version"`
	UpdateDate       string `json:"update_date"`
	RecUpdateAddress string `json:"rec_upd_address"`
	RecUpdateHouses  string `json:"rec_upd_houses"`
}
