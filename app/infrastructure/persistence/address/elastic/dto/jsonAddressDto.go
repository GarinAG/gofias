package dto

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/util"
	"gopkg.in/jeevatkm/go-model.v1"
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
	RegionGuid      string `json:"district_guid"`
	RegionKladr     string `json:"district_kladr"`
	Region          string `json:"district"`
	RegionType      string `json:"district_type"`
	RegionFull      string `json:"district_full"`
	AreaGuid        string `json:"area_guid"`
	AreaKladr       string `json:"area_kladr"`
	Area            string `json:"area"`
	AreaType        string `json:"area_type"`
	AreaFull        string `json:"area_full"`
	CityGuid        string `json:"city_guid"`
	CityKladr       string `json:"city_kladr"`
	City            string `json:"city"`
	CityType        string `json:"city_type"`
	CityFull        string `json:"city_full"`
	SettlementGuid  string `json:"settlement_guid"`
	SettlementKladr string `json:"settlement_kladr"`
	Settlement      string `json:"settlement"`
	SettlementType  string `json:"settlement_type"`
	SettlementFull  string `json:"settlement_full"`
	StreetGuid      string `json:"street_guid"`
	StreetKladr     string `json:"street_kladr"`
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
	address := entity.AddressObject{}
	model.Copy(&address, item)

	return &address
}

// Конвертирует объект адреса в объект адреса эластика
func (item *JsonAddressDto) GetFromEntity(entity entity.AddressObject) {
	model.Copy(item, entity)
	item.FormalName = strings.Trim(entity.FormalName, " -.,")
	item.ShortName = strings.Trim(entity.ShortName, " -.,")
	item.OffName = strings.Trim(entity.OffName, " -.,")
	if item.FullName == "" {
		item.FullName = util.PrepareFullName(item.ShortName, item.FormalName)
	}
	if item.FullAddress == "" {
		item.FullAddress = item.FullName
	}
	if item.AddressSuggest == "" {
		item.AddressSuggest = util.PrepareSuggest("", item.ShortName, item.FormalName)
	}

	item.UpdateBazisDate()
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
	item.BazisUpdateDate = time.Now().Format(util.TimeFormat)
}

// Заполняет объект адреса эластика из данных адреса
func (item *JsonAddressDto) UpdateFromExistItem(entity entity.AddressObject) {
	if entity.FullAddress != "" {
		item.FullAddress = entity.FullAddress
	}
	if entity.AddressSuggest != "" {
		item.AddressSuggest = entity.AddressSuggest
	}
	if entity.RegionGuid != "" {
		item.RegionGuid = entity.RegionGuid
	}
	if entity.Region != "" {
		item.Region = entity.Region
	}
	if entity.RegionType != "" {
		item.RegionType = entity.RegionType
	}
	if entity.RegionFull != "" {
		item.RegionFull = entity.RegionFull
	}
	if entity.AreaGuid != "" {
		item.AreaGuid = entity.AreaGuid
	}
	if entity.Area != "" {
		item.Area = entity.Area
	}
	if entity.AreaType != "" {
		item.AreaType = entity.AreaType
	}
	if entity.AreaFull != "" {
		item.AreaFull = entity.AreaFull
	}
	if entity.CityGuid != "" {
		item.CityGuid = entity.CityGuid
	}
	if entity.City != "" {
		item.City = entity.City
	}
	if entity.CityType != "" {
		item.CityType = entity.CityType
	}
	if entity.CityFull != "" {
		item.CityFull = entity.CityFull
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
