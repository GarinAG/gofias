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
	FullAddress     string `json:"full_address"`
	AddressSuggest  string `json:"address_suggest"`
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
	Location        string `json:"location"`
	BazisUpdateDate string `json:"bazis_update_date"`
}

func (item *JsonHouseDto) ToEntity() *entity.HouseObject {
	return &entity.HouseObject{
		ID:             item.ID,
		HouseGuid:      item.HouseGuid,
		AoGuid:         item.AoGuid,
		HouseNum:       item.HouseNum,
		HouseFullNum:   item.HouseFullNum,
		FullAddress:    item.FullAddress,
		AddressSuggest: item.AddressSuggest,
		PostalCode:     item.PostalCode,
		Okato:          item.Okato,
		Oktmo:          item.Oktmo,
		StartDate:      item.StartDate,
		EndDate:        item.EndDate,
		UpdateDate:     item.UpdateDate,
		DivType:        item.DivType,
		BuildNum:       item.BuildNum,
		StructNum:      item.StructNum,
		Counter:        item.Counter,
		CadNum:         item.CadNum,
		Location:       item.Location,
	}
}

func (item *JsonHouseDto) GetFromEntity(entity entity.HouseObject) {
	if entity.HouseFullNum == "" {
		fullNum := "д. " + entity.HouseNum
		if entity.StructNum != "" {
			fullNum += ", стр. " + entity.StructNum
		}
		if entity.BuildNum != "" {
			fullNum += ", кор. " + entity.BuildNum
		}

		entity.HouseFullNum = fullNum
	}

	if entity.AddressSuggest == "" {
		suggest := "дом (д.) " + entity.HouseNum
		if entity.StructNum != "" {
			suggest += ", строение (стр.) " + entity.StructNum
		}
		if entity.BuildNum != "" {
			suggest += ", корпус (кор.) " + entity.BuildNum
		}

		entity.AddressSuggest = suggest
	}

	item.ID = entity.ID
	item.HouseGuid = entity.HouseGuid
	item.AoGuid = entity.AoGuid
	item.HouseNum = entity.HouseNum
	item.HouseFullNum = entity.HouseFullNum
	item.FullAddress = entity.FullAddress
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
	item.Location = entity.Location
	item.BazisUpdateDate = time.Now().Format("2006-01-02") + "T00:00:00Z"
}

func (item *JsonHouseDto) IsActive() bool {
	end, err := time.Parse("2006-01-02", item.EndDate)
	if err != nil || end.Unix() <= time.Now().Unix() {
		return false
	}

	return true
}

func (item *JsonHouseDto) UpdateBazisDate() {
	item.BazisUpdateDate = time.Now().Format("2006-01-02") + "T00:00:00Z"
}
