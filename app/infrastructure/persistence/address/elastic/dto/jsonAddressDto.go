package dto

import (
	"github.com/GarinAG/gofias/domain/address/entity"
	"time"
)

type JsonAddressDto struct {
	ID                   string         `json:"_id"`
	AoGuid               string         `json:"ao_guid"`
	ParentGuid           string         `json:"parent_guid"`
	FormalName           string         `json:"formal_name"`
	ShortName            string         `json:"short_name"`
	AoLevel              string         `json:"ao_level"`
	OffName              string         `json:"off_name"`
	AreaCode             string         `json:"area_code"`
	CityCode             string         `json:"city_code"`
	PlaceCode            string         `json:"place_code"`
	AutoCode             string         `json:"auto_code"`
	PlanCode             string         `json:"plan_code"`
	StreetCode           string         `json:"street_code"`
	CTarCode             string         `json:"city_ar_code"`
	ExtrCode             string         `json:"extr_code"`
	SextCode             string         `json:"sub_ext_code"`
	Code                 string         `json:"code"`
	RegionCode           string         `json:"region_code"`
	PlainCode            string         `json:"plain_code"`
	PostalCode           string         `json:"postal_code"`
	Okato                string         `json:"okato"`
	Oktmo                string         `json:"oktmo"`
	IfNsFl               string         `json:"ifns_fl"`
	IfNsUl               string         `json:"ifns_ul"`
	TerrIfNsFl           string         `json:"terr_ifns_fl"`
	TerrIfNsUl           string         `json:"terr_ifns_ul"`
	NormDoc              string         `json:"norm_doc"`
	ActStatus            string         `json:"act_status"`
	LiveStatus           string         `json:"live_status"`
	CurrStatus           string         `json:"curr_status"`
	OperStatus           string         `json:"oper_status"`
	StartDate            string         `json:"start_date"`
	EndDate              string         `json:"end_date"`
	UpdateDate           string         `json:"update_date"`
	StreetType           string         `json:"street_type"`
	Street               string         `json:"street"`
	Settlement           string         `json:"settlement"`
	SettlementType       string         `json:"settlement_type"`
	District             string         `json:"district"`
	DistrictType         string         `json:"district_type"`
	StreetAddressSuggest string         `json:"street_address_suggest"`
	FullAddress          string         `json:"full_address"`
	Houses               []JsonHouseDto `json:"houses"`
	BazisCreateDate      string         `json:"bazis_create_date"`
	BazisUpdateDate      string         `json:"bazis_update_date"`
	BazisFinishDate      string         `json:"bazis_finish_date"`
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
		AreaCode:   item.AreaCode,
		CityCode:   item.CityCode,
		PlaceCode:  item.PlaceCode,
		AutoCode:   item.AutoCode,
		PlanCode:   item.PlanCode,
		StreetCode: item.StreetCode,
		CTarCode:   item.CTarCode,
		ExtrCode:   item.ExtrCode,
		SextCode:   item.SextCode,
		Code:       item.Code,
		RegionCode: item.RegionCode,
		PlainCode:  item.PlainCode,
		PostalCode: item.PostalCode,
		Okato:      item.Okato,
		Oktmo:      item.Oktmo,
		IfNsFl:     item.IfNsFl,
		IfNsUl:     item.IfNsUl,
		TerrIfNsFl: item.TerrIfNsFl,
		TerrIfNsUl: item.TerrIfNsUl,
		NormDoc:    item.NormDoc,
		ActStatus:  item.ActStatus,
		LiveStatus: item.LiveStatus,
		CurrStatus: item.CurrStatus,
		OperStatus: item.OperStatus,
		StartDate:  item.StartDate,
		EndDate:    item.EndDate,
		UpdateDate: item.UpdateDate,
	}
}

func (item *JsonAddressDto) GetFromEntity(entity entity.AddressObject) {
	item.ID = entity.ID
	item.AoGuid = entity.AoGuid
	item.ParentGuid = entity.ParentGuid
	item.FormalName = entity.FormalName
	item.ShortName = entity.ShortName
	item.AoLevel = entity.AoLevel
	item.OffName = entity.OffName
	item.AreaCode = entity.AreaCode
	item.CityCode = entity.CityCode
	item.PlaceCode = entity.PlaceCode
	item.AutoCode = entity.AutoCode
	item.PlanCode = entity.PlanCode
	item.StreetCode = entity.StreetCode
	item.CTarCode = entity.CTarCode
	item.ExtrCode = entity.ExtrCode
	item.SextCode = entity.SextCode
	item.Code = entity.Code
	item.RegionCode = entity.RegionCode
	item.PlainCode = entity.PlainCode
	item.PostalCode = entity.PostalCode
	item.Okato = entity.Okato
	item.Oktmo = entity.Oktmo
	item.IfNsFl = entity.IfNsFl
	item.IfNsUl = entity.IfNsUl
	item.TerrIfNsFl = entity.TerrIfNsFl
	item.TerrIfNsUl = entity.TerrIfNsUl
	item.NormDoc = entity.NormDoc
	item.ActStatus = entity.ActStatus
	item.LiveStatus = entity.LiveStatus
	item.CurrStatus = entity.CurrStatus
	item.OperStatus = entity.OperStatus
	item.StartDate = entity.StartDate
	item.EndDate = entity.EndDate
	item.UpdateDate = entity.UpdateDate
	item.BazisUpdateDate = time.Now().Format("2006-01-02") + "T00:00:00Z"
	item.BazisFinishDate = entity.EndDate
}
