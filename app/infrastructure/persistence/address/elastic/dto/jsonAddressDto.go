package dto

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/util"
	"strings"
	"time"
)

type JsonAddressDto struct {
	ID              string                `json:"ao_id"`
	AoGuid          string                `json:"ao_guid"`
	ParentGuid      string                `json:"parent_guid"`
	FormalName      string                `json:"formal_name"`
	ShortName       string                `json:"short_name"`
	AoLevel         int                   `json:"ao_level"`
	OffName         string                `json:"off_name"`
	FullName        string                `json:"full_name"`
	Code            string                `json:"code"`
	RegionCode      string                `json:"region_code"`
	PostalCode      string                `json:"postal_code"`
	Okato           string                `json:"okato"`
	Oktmo           string                `json:"oktmo"`
	ActStatus       string                `json:"act_status"`
	LiveStatus      string                `json:"live_status"`
	CurrStatus      string                `json:"curr_status"`
	StartDate       string                `json:"start_date"`
	EndDate         string                `json:"end_date"`
	UpdateDate      string                `json:"update_date"`
	District        string                `json:"district"`
	DistrictType    string                `json:"district_type"`
	DistrictFull    string                `json:"district_full"`
	Settlement      string                `json:"settlement"`
	SettlementType  string                `json:"settlement_type"`
	SettlementFull  string                `json:"settlement_full"`
	Street          string                `json:"street"`
	StreetType      string                `json:"street_type"`
	StreetFull      string                `json:"street_full"`
	AddressSuggest  string                `json:"address_suggest"`
	FullAddress     string                `json:"full_address"`
	Houses          []JsonAddressHouseDto `json:"houses"`
	BazisUpdateDate string                `json:"bazis_update_date"`
}

func (item *JsonAddressDto) ToEntity() *entity.AddressObject {
	return &entity.AddressObject{
		ID:         item.ID,
		AoGuid:     item.AoGuid,
		ParentGuid: item.ParentGuid,
		FormalName: item.FormalName,
		ShortName:  item.ShortName,
		AoLevel:    item.AoLevel,
		OffName:    item.OffName,
		Code:       item.Code,
		RegionCode: item.RegionCode,
		PostalCode: item.PostalCode,
		Okato:      item.Okato,
		Oktmo:      item.Oktmo,
		ActStatus:  item.ActStatus,
		LiveStatus: item.LiveStatus,
		CurrStatus: item.CurrStatus,
		StartDate:  item.StartDate,
		EndDate:    item.EndDate,
		UpdateDate: item.UpdateDate,
	}
}

func (item *JsonAddressDto) GetFromEntity(entity entity.AddressObject) {
	item.ID = entity.ID
	item.AoGuid = entity.AoGuid
	item.ParentGuid = entity.ParentGuid
	item.AoLevel = entity.AoLevel
	item.FormalName = strings.Trim(entity.FormalName, " -.,")
	item.ShortName = strings.Trim(entity.ShortName, " -.,")
	item.OffName = strings.Trim(entity.OffName, " -.,")
	item.FullName = util.PrepareFullName(item.ShortName, item.OffName)
	item.FullAddress = item.FullName
	item.AddressSuggest = strings.ToLower(strings.TrimSpace(item.OffName))
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
	item.BazisUpdateDate = time.Now().Format("2006-01-02") + "T00:00:00Z"
}
