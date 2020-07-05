package main

import "time"

const (
	houseIndexSettings = `
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
		  }
		}
	  },
	  "mappings": {
		"dynamic": false,
		"properties": {
		  "ao_guid": {
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
	`
	houseDropPipeline = `
	{
	  "description": "drop old houses",
	  "processors": [
		{
		  "drop": {
			"if": "ZonedDateTime zdt = ZonedDateTime.parse(ctx.bazis_update_date); long millisDateTime = zdt.toInstant().toEpochMilli(); ZonedDateTime nowDate = ZonedDateTime.ofInstant(Instant.ofEpochMilli(millisDateTime), ZoneId.of('Z')); ZonedDateTime endDateZDT = ZonedDateTime.parse(ctx.end_date + 'T00:00:00Z'); long millisDateTimeEndDate = endDateZDT.toInstant().toEpochMilli(); ZonedDateTime endDate = ZonedDateTime.ofInstant(Instant.ofEpochMilli(millisDateTimeEndDate), ZoneId.of('Z')); return endDate.isBefore(nowDate);"
		  }
		}
	  ]
	}
	`
)

type HouseItemElastic struct {
	ID              string `json:"_id"`
	AoGuid          string `json:"ao_guid"`
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
	HouseNum        string `json:"house_num"`
	BuildNum        string `json:"build_num"`
	StructNum       string `json:"str_num"`
	Counter         string `json:"counter"`
	CadNum          string `json:"cad_num"`
	BazisCreateDate string `json:"bazis_create_date"`
	BazisUpdateDate string `json:"bazis_update_date"`
	BazisFinishDate string `json:"bazis_finish_date"`
}

func getHouseElasticStructFromXml(item HouseItem) HouseItemElastic {
	currentTime := time.Now().Format("2006-01-02") + dateTimeZone
	saveTime := currentTime
	if isUpdate {
		saveTime = versionDate
	}

	return HouseItemElastic{
		ID:              item.HouseId,
		AoGuid:          item.AoGuid,
		RegionCode:      item.RegionCode,
		PostalCode:      item.PostalCode,
		Okato:           item.Okato,
		Oktmo:           item.Oktmo,
		IfNsFl:          item.IfNsFl,
		IfNsUl:          item.IfNsUl,
		TerrIfNsFl:      item.TerrIfNsFl,
		TerrIfNsUl:      item.TerrIfNsUl,
		NormDoc:         item.NormDoc,
		StartDate:       item.StartDate,
		EndDate:         item.EndDate,
		UpdateDate:      item.UpdateDate,
		DivType:         item.DivType,
		HouseNum:        item.HouseNum,
		BuildNum:        item.BuildNum,
		StructNum:       item.StructNum,
		Counter:         item.Counter,
		CadNum:          item.CadNum,
		BazisCreateDate: currentTime,
		BazisUpdateDate: saveTime,
		BazisFinishDate: item.EndDate,
	}
}
