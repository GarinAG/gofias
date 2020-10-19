package dto

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/util"
	"strings"
	"time"
)

// Объект адреса в эластике
type JsonAddressDto struct {
	ID              string `json:"ao_id"`
	AoGuid          string `json:"ao_guid"`
	ParentGuid      string `json:"parent_guid"`
	FormalName      string `json:"formal_name"`
	ShortName       string `json:"short_name"`
	AoLevel         int    `json:"ao_level"`
	OffName         string `json:"off_name"`
	FullName        string `json:"full_name"`
	Code            string `json:"code"`
	RegionCode      string `json:"region_code"`
	PostalCode      string `json:"postal_code"`
	Okato           string `json:"okato"`
	Oktmo           string `json:"oktmo"`
	ActStatus       string `json:"act_status"`
	LiveStatus      string `json:"live_status"`
	CurrStatus      string `json:"curr_status"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	UpdateDate      string `json:"update_date"`
	DistrictGuid    string `json:"district_guid"`
	District        string `json:"district"`
	DistrictType    string `json:"district_type"`
	DistrictFull    string `json:"district_full"`
	SettlementGuid  string `json:"settlement_guid"`
	Settlement      string `json:"settlement"`
	SettlementType  string `json:"settlement_type"`
	SettlementFull  string `json:"settlement_full"`
	Street          string `json:"street"`
	StreetType      string `json:"street_type"`
	StreetFull      string `json:"street_full"`
	AddressSuggest  string `json:"address_suggest"`
	FullAddress     string `json:"full_address"`
	Location        string `json:"location"`
	BazisUpdateDate string `json:"bazis_update_date"`
}

// Конвертирует объект адреса эластика в объект адрес
func (item *JsonAddressDto) ToEntity() *entity.AddressObject {
	return &entity.AddressObject{
		ID:             item.ID,
		AoGuid:         item.AoGuid,
		ParentGuid:     item.ParentGuid,
		FormalName:     item.FormalName,
		ShortName:      item.ShortName,
		AoLevel:        item.AoLevel,
		OffName:        item.OffName,
		Code:           item.Code,
		RegionCode:     item.RegionCode,
		PostalCode:     item.PostalCode,
		Okato:          item.Okato,
		Oktmo:          item.Oktmo,
		ActStatus:      item.ActStatus,
		LiveStatus:     item.LiveStatus,
		CurrStatus:     item.CurrStatus,
		StartDate:      item.StartDate,
		EndDate:        item.EndDate,
		UpdateDate:     item.UpdateDate,
		FullName:       item.FullName,
		FullAddress:    item.FullAddress,
		AddressSuggest: item.AddressSuggest,
		DistrictGuid:   item.DistrictGuid,
		District:       item.District,
		DistrictType:   item.DistrictType,
		DistrictFull:   item.DistrictFull,
		SettlementGuid: item.SettlementGuid,
		Settlement:     item.Settlement,
		SettlementType: item.SettlementType,
		SettlementFull: item.SettlementFull,
		Street:         item.Street,
		StreetType:     item.StreetType,
		StreetFull:     item.StreetFull,
		Location:       item.Location,
	}
}

// Конвертирует объект адреса в объект адреса эластика
func (item *JsonAddressDto) GetFromEntity(entity entity.AddressObject) {
	item.ID = entity.ID
	item.AoGuid = entity.AoGuid
	item.ParentGuid = entity.ParentGuid
	item.AoLevel = entity.AoLevel
	item.FormalName = strings.Trim(entity.FormalName, " -.,")
	item.ShortName = strings.Trim(entity.ShortName, " -.,")
	item.OffName = strings.Trim(entity.OffName, " -.,")
	item.FullName = entity.FullName
	item.FullAddress = entity.FullAddress
	item.AddressSuggest = entity.AddressSuggest
	if item.FullAddress == "" {
		item.FullAddress = item.FullName
	}
	if item.FullName == "" {
		item.FullName = util.PrepareFullName(item.ShortName, item.FormalName)
	}
	if item.AddressSuggest == "" {
		item.AddressSuggest = util.PrepareSuggest("", item.ShortName, item.FormalName)
	}

	item.Code = entity.Code
	item.RegionCode = entity.RegionCode
	item.PostalCode = entity.PostalCode
	item.Okato = entity.Okato
	item.Oktmo = entity.Oktmo
	item.ActStatus = entity.ActStatus
	item.LiveStatus = entity.LiveStatus
	item.CurrStatus = entity.CurrStatus
	item.StartDate = entity.StartDate
	item.EndDate = entity.EndDate
	item.UpdateDate = entity.UpdateDate
	item.DistrictGuid = entity.DistrictGuid
	item.District = entity.District
	item.DistrictType = entity.DistrictType
	item.DistrictFull = entity.DistrictFull
	item.SettlementGuid = entity.SettlementGuid
	item.Settlement = entity.Settlement
	item.SettlementType = entity.SettlementType
	item.SettlementFull = entity.SettlementFull
	item.Street = entity.Street
	item.StreetType = entity.StreetType
	item.StreetFull = entity.StreetFull
	item.Location = entity.Location
	item.BazisUpdateDate = time.Now().Format("2006-01-02") + "T00:00:00Z"
}

// Проверяет активность объекта
func (item *JsonAddressDto) IsActive() bool {
	if item.CurrStatus != "0" ||
		item.ActStatus != "1" ||
		item.LiveStatus != "1" {

		return false
	}

	return true
}

// Устанавливает время обновления объекта
func (item *JsonAddressDto) UpdateBazisDate() {
	item.BazisUpdateDate = time.Now().Format("2006-01-02") + "T00:00:00Z"
}

// Заполняет объект адреса эластика из данных адреса
func (item *JsonAddressDto) UpdateFromExistItem(entity entity.AddressObject) {
	if entity.FullAddress != "" {
		item.FullAddress = entity.FullName
	}
	if entity.AddressSuggest != "" {
		item.AddressSuggest = entity.AddressSuggest
	}
	if entity.DistrictGuid != "" {
		item.DistrictGuid = entity.DistrictGuid
	}
	if entity.District != "" {
		item.District = entity.District
	}
	if entity.DistrictType != "" {
		item.DistrictType = entity.DistrictType
	}
	if entity.DistrictFull != "" {
		item.DistrictFull = entity.DistrictFull
	}
	if entity.SettlementGuid != "" {
		item.SettlementGuid = entity.SettlementGuid
	}
	if entity.Settlement != "" {
		item.Settlement = entity.Settlement
	}
	if entity.SettlementType != "" {
		item.SettlementType = entity.SettlementType
	}
	if entity.SettlementFull != "" {
		item.SettlementFull = entity.SettlementFull
	}
	if entity.Street != "" {
		item.Street = entity.Street
	}
	if entity.StreetType != "" {
		item.StreetType = entity.StreetType
	}
	if entity.StreetFull != "" {
		item.StreetFull = entity.StreetFull
	}
	if entity.Location != "" {
		item.Location = entity.Location
	}
}
