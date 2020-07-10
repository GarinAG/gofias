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
	housesPipelineId  = "house_drop_pipeline"
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

type ElasticHouseRepository struct {
	elasticClient *elasticHelper.Client
	indexName     string
}

func NewElasticHouseRepository(elasticClient *elasticHelper.Client, configInterface interfaces.ConfigInterface) repository.HouseRepositoryInterface {
	return &ElasticHouseRepository{
		elasticClient: elasticClient,
		indexName:     configInterface.GetString("project.prefix") + entity.HouseObject{}.TableName(),
	}
}

func (a *ElasticHouseRepository) Init() error {
	err := a.elasticClient.CreateIndex(a.indexName, houseIndexSettings)
	if err != nil {
		return err
	}

	return a.elasticClient.CreatePreprocessor(housesPipelineId, houseDropPipeline)
}

func (a *ElasticHouseRepository) Clear() error {
	return a.elasticClient.DropIndex(a.indexName)
}

func (a *ElasticHouseRepository) GetByAddressGuid(guid string) (*entity.HouseObject, error) {
	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewTermQuery("ao_guid", guid)).
		Size(1).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	var item *entity.HouseObject
	if len(res.Hits.Hits) > 0 {
		if err := json.Unmarshal(res.Hits.Hits[0].Source, &item); err != nil {
			return nil, err
		}

		return item, nil
	}

	return nil, nil
}

func (a *ElasticHouseRepository) InsertUpdateCollection(collection []interface{}, isFull bool) error {
	bulk := a.elasticClient.Client.Bulk().Index(a.indexName).Pipeline(housesPipelineId)
	for _, item := range collection {
		item := item.(dto.JsonHouseDto)
		bulk.Add(elastic.NewBulkIndexRequest().Id(item.ID).Doc(item))
	}

	if bulk.NumberOfActions() > 0 {
		// Commit
		res, err := bulk.Do(context.Background())
		if err != nil {
			return err
		}
		if res.Errors {
			return errors.New("Add houses bulk commit failed")
		}
	}

	return nil
}

func (a *ElasticHouseRepository) Flush(fool bool, params ...interface{}) error {

	return nil
}

func (a *ElasticHouseRepository) convertToEntity(item dto.JsonHouseDto) entity.HouseObject {
	return entity.HouseObject{
		ID:         item.ID,
		AoGuid:     item.AoGuid,
		HouseNum:   item.HouseNum,
		RegionCode: item.RegionCode,
		PostalCode: item.PostalCode,
		Okato:      item.Okato,
		Oktmo:      item.Oktmo,
		IfNsFl:     item.IfNsFl,
		IfNsUl:     item.IfNsUl,
		TerrIfNsFl: item.TerrIfNsFl,
		TerrIfNsUl: item.TerrIfNsUl,
		NormDoc:    item.NormDoc,
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

func (a *ElasticHouseRepository) convertToDto(item entity.HouseObject) dto.JsonHouseDto {
	return dto.JsonHouseDto{
		ID:              item.ID,
		AoGuid:          item.AoGuid,
		HouseNum:        item.HouseNum,
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
		BuildNum:        item.BuildNum,
		StructNum:       item.StructNum,
		Counter:         item.Counter,
		CadNum:          item.CadNum,
		BazisUpdateDate: time.Now().Format("2006-01-02") + "T00:00:00Z",
		BazisFinishDate: item.EndDate,
	}
}
