package dto

type JsonHouseDto struct {
	ID              string `json:"_id"`
	AoGuid          string `json:"ao_guid"`
	HouseNum        string `json:"house_num"`
	RegionCode      string `json:"region_code"`
	PostalCode      string `json:"postal_code"`
	Okato           string `json:"okato"`
	Oktmo           string `json:"oktmo"`
	IfNsFl          string `json:"ifns_fl"`
	IfNsUl          string `json:"ifns_ul"`
	TerrIfNsFl      string `json:"terr_ifns_fl"`
	TerrIfNsUl      string `json:"terr_ifns_ul"`
	NormDoc         string `json:"norm_doc"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	UpdateDate      string `json:"update_date"`
	DivType         string `json:"div_type"`
	BuildNum        string `json:"build_num"`
	StructNum       string `json:"str_num"`
	Counter         string `json:"counter"`
	CadNum          string `json:"cad_num"`
	BazisCreateDate string `json:"bazis_create_date"`
	BazisUpdateDate string `json:"bazis_update_date"`
	BazisFinishDate string `json:"bazis_finish_date"`
}
