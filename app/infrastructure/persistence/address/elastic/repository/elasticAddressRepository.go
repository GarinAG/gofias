package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/domain/address/repository"
	"github.com/GarinAG/gofias/infrastructure/persistence/address/elastic/dto"
	elasticHelper "github.com/GarinAG/gofias/infrastructure/persistence/elastic"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/olivere/elastic/v7"
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
			"type": "text",
			"analyzer": "autocomplete",
			"search_analyzer": "stop_analyzer"
		  },
		  "district_full": {
			"type": "text",
			"analyzer": "autocomplete",
			"search_analyzer": "stop_analyzer"
		  },
		  "settlement_full": {
			"type": "text",
			"analyzer": "autocomplete",
			"search_analyzer": "stop_analyzer"
		  },
		  "street_full": {
			"type": "text",
			"analyzer": "autocomplete",
			"search_analyzer": "stop_analyzer"
		  },
		  "formal_name": {
			"type": "text",
			"analyzer": "autocomplete",
			"search_analyzer": "stop_analyzer"
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
				"type": "text",
				"analyzer": "autocomplete",
				"search_analyzer": "stop_analyzer"
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
	addrPipelineId   = "addr_drop_pipeline"
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

type ElasticAddressRepository struct {
	elasticClient *elasticHelper.Client
	indexName     string
}

func NewElasticAddressRepository(elasticClient *elasticHelper.Client, configInterface interfaces.ConfigInterface) repository.AddressRepositoryInterface {
	return &ElasticAddressRepository{
		elasticClient: elasticClient,
		indexName:     configInterface.GetString("project.prefix") + entity.AddressObject{}.TableName(),
	}
}

func (a *ElasticAddressRepository) Init() error {
	err := a.elasticClient.CreateIndex(a.indexName, addrIndexSettings)
	if err != nil {
		return err
	}

	return a.elasticClient.CreatePreprocessor(addrPipelineId, addrDropPipeline)
}

func (a *ElasticAddressRepository) Clear() error {
	return a.elasticClient.DropIndex(a.indexName)
}

func (a *ElasticAddressRepository) GetByFormalName(term string) (*entity.AddressObject, error) {
	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewMatchQuery("formal_name", term)).
		Size(1).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	var item *entity.AddressObject
	if len(res.Hits.Hits) > 0 {
		if err := json.Unmarshal(res.Hits.Hits[0].Source, &item); err != nil {
			return nil, err
		}

		return item, nil
	}

	return nil, nil
}

func (a *ElasticAddressRepository) GetByGuid(guid string) (*entity.AddressObject, error) {
	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewTermQuery("ao_guid", guid)).
		Size(1).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	var item *entity.AddressObject
	if len(res.Hits.Hits) > 0 {
		if err := json.Unmarshal(res.Hits.Hits[0].Source, &item); err != nil {
			return nil, err
		}

		return item, nil
	}

	return nil, nil
}

func (a ElasticAddressRepository) GetCityByFormalName(term string) (*entity.AddressObject, error) {
	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewBoolQuery().Filter(
			elastic.NewTermQuery("short_name", "г"),
			elastic.NewTermsQuery("ao_level", 1, 4)).
			Must(elastic.NewMatchQuery("formal_name", term))).
		Sort("ao_level", true).
		Size(1).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	var item *entity.AddressObject
	if len(res.Hits.Hits) > 0 {
		if err := json.Unmarshal(res.Hits.Hits[0].Source, &item); err != nil {
			return nil, err
		}

		return item, nil
	}

	return nil, nil
}

func (a *ElasticAddressRepository) GetCities() ([]*entity.AddressObject, error) {
	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewBoolQuery().Filter(
			elastic.NewTermQuery("short_name", "г"),
			elastic.NewTermsQuery("ao_level", 1, 4))).
		Sort("ao_level", true).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	var items []*entity.AddressObject
	var item *entity.AddressObject
	if len(res.Hits.Hits) > 0 {
		for _, el := range res.Hits.Hits {
			if err := json.Unmarshal(el.Source, &item); err != nil {
				return nil, err
			}
		}
		items = append(items, item)
	}

	return items, nil
}

func (a *ElasticAddressRepository) GetCitiesByTerm(term string, count int64) ([]*entity.AddressObject, error) {
	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewBoolQuery().Filter(
			elastic.NewTermQuery("short_name", "г"),
			elastic.NewTermQuery("formal_name", term),
			elastic.NewTermsQuery("ao_level", 1, 4))).
		Sort("ao_level", true).
		Size(int(count)).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	var items []*entity.AddressObject
	var item *entity.AddressObject
	if len(res.Hits.Hits) > 0 {
		for _, el := range res.Hits.Hits {
			if err := json.Unmarshal(el.Source, &item); err != nil {
				return nil, err
			}
		}
		items = append(items, item)
	}

	return items, nil
}

func (a *ElasticAddressRepository) InsertUpdateCollection(collection []interface{}, isFull bool) error {
	bulk := a.elasticClient.Client.Bulk().Index(a.indexName).Pipeline(addrPipelineId)
	for _, item := range collection {
		item := item.(dto.JsonAddressDto)
		bulk.Add(elastic.NewBulkIndexRequest().Id(item.ID).Doc(item))
	}

	if bulk.NumberOfActions() > 0 {
		// Commit
		res, err := bulk.Do(context.Background())
		if err != nil {
			return err
		}
		if res.Errors {
			return errors.New("Add addresses bulk commit failed")
		}
	}

	return nil
}

func (a *ElasticAddressRepository) Flush(fool bool, params ...interface{}) error {

	return nil
}

func (a *ElasticAddressRepository) convertToEntity(item dto.JsonAddressDto) entity.AddressObject {
	return entity.AddressObject{
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

func (a *ElasticAddressRepository) convertToDto(item entity.AddressObject) dto.JsonAddressDto {
	return dto.JsonAddressDto{
		ID:              item.ID,
		AoGuid:          item.AoGuid,
		ParentGuid:      item.ParentGuid,
		FormalName:      item.FormalName,
		ShortName:       item.ShortName,
		AoLevel:         item.AoLevel,
		OffName:         item.OffName,
		AreaCode:        item.AreaCode,
		CityCode:        item.CityCode,
		PlaceCode:       item.PlaceCode,
		AutoCode:        item.AutoCode,
		PlanCode:        item.PlanCode,
		StreetCode:      item.StreetCode,
		CTarCode:        item.CTarCode,
		ExtrCode:        item.ExtrCode,
		SextCode:        item.SextCode,
		Code:            item.Code,
		RegionCode:      item.RegionCode,
		PlainCode:       item.PlainCode,
		PostalCode:      item.PostalCode,
		Okato:           item.Okato,
		Oktmo:           item.Oktmo,
		IfNsFl:          item.IfNsFl,
		IfNsUl:          item.IfNsUl,
		TerrIfNsFl:      item.TerrIfNsFl,
		TerrIfNsUl:      item.TerrIfNsUl,
		NormDoc:         item.NormDoc,
		ActStatus:       item.ActStatus,
		LiveStatus:      item.LiveStatus,
		CurrStatus:      item.CurrStatus,
		OperStatus:      item.OperStatus,
		StartDate:       item.StartDate,
		EndDate:         item.EndDate,
		UpdateDate:      item.UpdateDate,
		BazisUpdateDate: time.Now().Format("2006-01-02") + "T00:00:00Z",
		BazisFinishDate: item.EndDate,
	}
}
