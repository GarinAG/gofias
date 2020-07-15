package dto

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"time"
)

type JsonHouseDto struct {
	ID              string `json:"house_id"`
	HouseGuid       string `json:"house_guid"`
	AoGuid          string `json:"ao_guid"`
	HouseNum        string `json:"house_num"`
	HouseFullNum    string `json:"house_full_num"`
	PostalCode      string `json:"postal_code"`
	Okato           string `json:"okato"`
	Oktmo           string `json:"oktmo"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	UpdateDate      string `json:"update_date"`
	DivType         string `json:"div_type"`
	BuildNum        string `json:"build_num"`
	StructNum       string `json:"str_num"`
	Counter         string `json:"counter"`
	CadNum          string `json:"cad_num"`
	BazisUpdateDate string `json:"bazis_update_date"`
}

func (item *JsonHouseDto) ToEntity() *entity.HouseObject {
	return &entity.HouseObject{
		ID:         item.ID,
		HouseGuid:  item.HouseGuid,
		AoGuid:     item.AoGuid,
		HouseNum:   item.HouseNum,
		PostalCode: item.PostalCode,
		Okato:      item.Okato,
		Oktmo:      item.Oktmo,
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
	fullNum := "д. " + entity.HouseNum
	if entity.StructNum != "" {
		fullNum += ", стр. " + entity.StructNum
	}
	if entity.BuildNum != "" {
		fullNum += ", кор. " + entity.BuildNum
	}

	item.ID = entity.ID
	item.HouseGuid = entity.HouseGuid
	item.AoGuid = entity.AoGuid
	item.HouseNum = entity.HouseNum
	item.HouseFullNum = fullNum
	item.PostalCode = entity.PostalCode
	item.Okato = entity.Okato
	item.Oktmo = entity.Oktmo
	item.StartDate = entity.StartDate
	item.EndDate = entity.EndDate
	item.UpdateDate = entity.UpdateDate
	item.DivType = entity.DivType
	item.BuildNum = entity.BuildNum
	item.StructNum = entity.StructNum
	item.Counter = entity.Counter
	item.CadNum = entity.CadNum
	item.BazisUpdateDate = time.Now().Format("2006-01-02") + "T00:00:00Z"
}
