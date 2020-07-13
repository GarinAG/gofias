package dto

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"time"
)

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

func (item *JsonHouseDto) ToEntity() *entity.HouseObject {
	return &entity.HouseObject{
		ID:         item.ID,
		AoGuid:     item.AoGuid,
		HouseNum:   item.HouseNum,
		RegionCode: item.RegionCode,
		PostalCode: item.PostalCode,
		Okato:      item.Okato,
		Oktmo:      item.Oktmo,
		IfNsFl:     item.IfNsFl,
		IfNsUl:     item.IfNsUl,
		TerrIfNsFl: item.TerrIfNsFl,
		TerrIfNsUl: item.TerrIfNsUl,
		NormDoc:    item.NormDoc,
		StartDate:  item.StartDate,
		EndDate:    item.EndDate,
		UpdateDate: item.UpdateDate,
		DivType:    item.DivType,
		BuildNum:   item.BuildNum,
		StructNum:  item.StructNum,
		Counter:    item.Counter,
		CadNum:     item.CadNum,
	}
}

func (item *JsonHouseDto) GetFromEntity(entity entity.HouseObject) {
	item.ID = entity.ID
	item.AoGuid = entity.AoGuid
	item.HouseNum = entity.HouseNum
	item.RegionCode = entity.RegionCode
	item.PostalCode = entity.PostalCode
	item.Okato = entity.Okato
	item.Oktmo = entity.Oktmo
	item.IfNsFl = entity.IfNsFl
	item.IfNsUl = entity.IfNsUl
	item.TerrIfNsFl = entity.TerrIfNsFl
	item.TerrIfNsUl = entity.TerrIfNsUl
	item.NormDoc = entity.NormDoc
	item.StartDate = entity.StartDate
	item.EndDate = entity.EndDate
	item.UpdateDate = entity.UpdateDate
	item.DivType = entity.DivType
	item.BuildNum = entity.BuildNum
	item.StructNum = entity.StructNum
	item.Counter = entity.Counter
	item.CadNum = entity.CadNum
	item.BazisUpdateDate = time.Now().Format("2006-01-02") + "T00:00:00Z"
	item.BazisFinishDate = entity.EndDate
}
