package gofias

import (
	"time"
)

const (
	addrIndexSettings = `
	{
	  "settings": {
		"index": {
		  "number_of_shards": 1,
		  "number_of_replicas": "0",
		  "refresh_interval": "-1",
		  "requests": {
			"cache": {
			  "enable": "true"
			}
		  },
		  "blocks": {
			"read_only_allow_delete": "false"
		  },
		  "analysis": {
			"filter": {
			  "autocomplete_filter": {
				"type": "edge_ngram",
				"min_gram": 2,
				"max_gram": 20
			  },
			  "fias_word_delimiter": {
				"type": "word_delimiter",
				"preserve_original": "true",
				"generate_word_parts": "false"
			  }
			},
			"analyzer": {
			  "autocomplete": {
				"type": "custom",
				"tokenizer": "standard",
				"filter": ["autocomplete_filter"]
			  },
			  "stop_analyzer": {
				"type": "custom",
				"tokenizer": "whitespace",
				"filter": ["lowercase", "fias_word_delimiter"]
			  }
			}
		  }
		}
	  },
	  "mappings": {
		"dynamic": false,
		"properties": {
		  "street_address_suggest": {
			"type": "text",
			"analyzer": "autocomplete",
			"search_analyzer": "stop_analyzer"
		  },
		  "full_address": {
			"type": "keyword"
		  },
		  "district_full": {
			"type": "keyword"
		  },
		  "settlement_full": {
			"type": "keyword"
		  },
		  "street_full": {
			"type": "keyword"
		  },
		  "formal_name": {
			"type": "keyword"
		  },
		  "short_name": {
			"type": "keyword"
		  },
		  "off_name": {
			"type": "keyword"
		  },
		  "curr_status": {
			"type": "integer"
		  },
		  "oper_status": {
			"type": "integer"
		  },
		  "act_status": {
			"type": "integer"
		  },
		  "live_status": {
			"type": "integer"
		  },
		  "cent_status": {
			"type": "integer"
		  },
		  "ao_guid": {
			"type": "keyword"
		  },
		  "parent_guid": {
			"type": "keyword"
		  },
		  "ao_level": {
			"type": "keyword"
		  },
		  "area_code": {
			"type": "keyword"
		  },
		  "auto_code": {
			"type": "keyword"
		  },
		  "city_ar_code": {
			"type": "keyword"
		  },
		  "city_code": {
			"type": "keyword"
		  },
		  "street_code": {
			"type": "keyword"
		  },
		  "extr_code": {
			"type": "keyword"
		  },
		  "sub_ext_code": {
			"type": "keyword"
		  },
		  "place_code": {
			"type": "keyword"
		  },
		  "plan_code": {
			"type": "keyword"
		  },
		  "plain_code": {
			"type": "keyword"
		  },
		  "code": {
			"type": "keyword"
		  },
		  "postal_code": {
			"type": "keyword"
		  },
		  "region_code": {
			"type": "keyword"
		  },
		  "street": {
			"type": "keyword"
		  },
		  "district": {
			"type": "keyword"
		  },
		  "district_type": {
			"type": "keyword"
		  },
		  "street_type": {
			"type": "keyword"
		  },
		  "settlement": {
			"type": "keyword"
		  },
		  "settlement_type": {
			"type": "keyword"
		  },
		  "okato": {
			"type": "keyword"
		  },
		  "oktmo": {
			"type": "keyword"
		  },
		  "ifns_fl": {
			"type": "keyword"
		  },
		  "ifns_ul": {
			"type": "keyword"
		  },
		  "terr_ifns_fl": {
			"type": "keyword"
		  },
		  "terr_ifns_ul": {
			"type": "keyword"
		  },
		  "norm_doc": {
			"type": "keyword"
		  },
		  "start_date": {
			"type": "date"
		  },
		  "end_date": {
			"type": "date"
		  },
		  "bazis_finish_date": {
			"type": "date"
		  },
		  "bazis_create_date": {
			"type": "date"
		  },
		  "bazis_update_date": {
			"type": "date"
		  },
		  "update_date": {
			"type": "date"
		  },
		  "location": {
			"type": "geo_point"
		  },
		  "houses": {
			"type": "nested",
			"properties": {
			  "houseId": {
				"type": "keyword"
			  },
			  "build_num": {
				"type": "keyword"
			  },
			  "house_num": {
				"type": "keyword"
			  },
			  "str_num": {
				"type": "keyword"
			  },
			  "ifns_fl": {
				"type": "keyword"
			  },
			  "ifns_ul": {
				"type": "keyword"
			  },
			  "postal_code": {
				"type": "keyword"
			  },
			  "counter": {
				"type": "keyword"
			  },
			  "end_date": {
				"type": "date"
			  },
			  "start_date": {
				"type": "date"
			  },
			  "update_date": {
				"type": "date"
			  },
			  "cad_num": {
				"type": "keyword"
			  },
			  "terr_ifns_fl": {
				"type": "keyword"
			  },
			  "terr_ifns_ul": {
				"type": "keyword"
			  },
			  "okato": {
				"type": "keyword"
			  },
			  "oktmo": {
				"type": "keyword"
			  },
			  "location": {
				"type": "geo_point"
			  }
			}
		  }
		}
	  }
	}
`
	addrDropPipeline = `
	{
	  "description":
	  "drop not actual addresses",
	  "processors": [{
		"drop": {
		  "if": "ctx.curr_status  != '0' "
		}
	  }, {
		"drop": {
		  "if": "ctx.act_status  != '1'"
		}
	  }, {
		"drop": {
		  "if": "ctx.live_status  != '1'"
		}
	  }]
	}
`
)

type AddressItemElastic struct {
	ID string `json:"_id"`
	AoGuid string `json:"ao_guid"`
	ParentGuid string `json:"parent_guid"`
	FormalName string `json:"formal_name"`
	OffName string `json:"off_name"`
	ShortName string `json:"short_name"`
	AoLevel string `json:"ao_level"`
	AreaCode string `json:"area_code"`
	CityCode string `json:"city_code"`
	PlaceCode string `json:"place_code"`
	AutoCode string `json:"auto_code"`
	PlanCode string `json:"plan_code"`
	StreetCode string `json:"street_code"`
	CTarCode string `json:"city_ar_code"`
	ExtrCode string `json:"extr_code"`
	SextCode string `json:"sub_ext_code"`
	Code string `json:"code"`
	RegionCode string `json:"region_code"`
	PlainCode string `json:"plain_code"`
	PostalCode string `json:"postal_code"`
	Okato string `json:"okato"`
	Oktmo string `json:"oktmo"`
	IfNsFl string `json:"ifns_fl"`
	IfNsUl string `json:"ifns_ul"`
	TerrIfNsFl string `json:"terr_ifns_fl"`
	TerrIfNsUl string `json:"terr_ifns_ul"`
	NormDoc string `json:"norm_doc"`
	ActStatus string `json:"act_status"`
	LiveStatus string `json:"live_status"`
	CurrStatus string `json:"curr_status"`
	OperStatus string `json:"oper_status"`
	StartDate string `json:"start_date"`
	EndDate string `json:"end_date"`
	UpdateDate string `json:"update_date"`
	StreetType string `json:"street_type"`
	Street string `json:"street"`
	Settlement string `json:"settlement"`
	SettlementType string `json:"settlement_type"`
	District string `json:"district"`
	DistrictType string `json:"district_type"`
	StreetAddressSuggest string `json:"street_address_suggest"`
	FullAddress string `json:"full_address"`
	Houses []HouseItemElastic `json:"houses"`
	BazisCreateDate string `json:"bazis_create_date"`
	BazisUpdateDate string `json:"bazis_update_date"`
	BazisFinishDate string `json:"bazis_finish_date"`
}

func getAddressElasticStructFromXml(item AddressItem) AddressItemElastic {
	currentTime := time.Now().Format("2006-01-02") + dateTimeZone
	saveTime := currentTime
	if isUpdate {
		saveTime = versionDate
	}

	return AddressItemElastic{
		ID: item.AoId,
		AoGuid: item.AoGuid,
		ParentGuid: item.ParentGuid,
		FormalName: item.FormalName,
		OffName: item.OffName,
		ShortName: item.ShortName,
		AoLevel: item.AoLevel,
		AreaCode: item.AreaCode,
		CityCode: item.CityCode,
		PlaceCode: item.PlaceCode,
		AutoCode: item.AutoCode,
		PlanCode: item.PlanCode,
		StreetCode: item.StreetCode,
		ExtrCode: item.ExtrCode,
		SextCode: item.SextCode,
		Code: item.Code,
		RegionCode: item.RegionCode,
		PlainCode: item.PlainCode,
		PostalCode: item.PostalCode,
		Okato: item.Okato,
		Oktmo: item.Oktmo,
		IfNsFl: item.IfNsFl,
		IfNsUl: item.IfNsUl,
		TerrIfNsFl: item.TerrIfNsFl,
		TerrIfNsUl: item.TerrIfNsUl,
		NormDoc: item.NormDoc,
		ActStatus: item.ActStatus,
		LiveStatus: item.LiveStatus,
		CurrStatus: item.CurrStatus,
		OperStatus: item.OperStatus,
		StartDate: item.StartDate,
		EndDate: item.EndDate,
		UpdateDate: item.UpdateDate,
		BazisCreateDate: currentTime,
		BazisUpdateDate: saveTime,
		BazisFinishDate: item.EndDate,
	}
}