package dto

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/util"
	"gopkg.in/jeevatkm/go-model.v1"
	"time"
)

// Объект дома в эластике
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

// Конвертирует объект дома эластика в объект дома
func (item *JsonHouseDto) ToEntity() *entity.HouseObject {
	house := entity.HouseObject{}
	model.Copy(&house, item)

	return &house
}

// Конвертирует объект дома в объект дома эластика
func (item *JsonHouseDto) GetFromEntity(entity entity.HouseObject) {
	model.Copy(item, entity)

	if item.HouseFullNum == "" {
		fullNum := "д. " + entity.HouseNum
		if entity.StructNum != "" {
			fullNum += ", стр. " + entity.StructNum
		}
		if entity.BuildNum != "" {
			fullNum += ", кор. " + entity.BuildNum
		}

		item.HouseFullNum = fullNum
	}
	if item.AddressSuggest == "" {
		suggest := "дом (д.) " + entity.HouseNum
		if entity.StructNum != "" {
			suggest += ", строение (стр.) " + entity.StructNum
		}
		if entity.BuildNum != "" {
			suggest += ", корпус (кор.) " + entity.BuildNum
		}

		item.AddressSuggest = suggest
	}
	if item.FullAddress == "" {
		item.FullAddress = item.HouseFullNum
	}

	item.UpdateBazisDate()
}

// Проверяет активность объекта
func (item *JsonHouseDto) IsActive() bool {
	end, err := time.Parse("2006-01-02", item.EndDate)
	if err != nil || end.Unix() <= time.Now().Unix() {
		return false
	}

	return true
}

// Устанавливает время обновления объекта
func (item *JsonHouseDto) UpdateBazisDate() {
	item.BazisUpdateDate = time.Now().Format(util.TimeFormat)
}

// Заполняет объект дома эластика из данных дома
func (item *JsonHouseDto) UpdateFromExistItem(entity entity.HouseObject) {
	if entity.FullAddress != "" {
		item.FullAddress = entity.FullAddress
	}
	if entity.AddressSuggest != "" {
		item.AddressSuggest = entity.AddressSuggest
	}
	if entity.Location != "" {
		item.Location = entity.Location
	}
}
