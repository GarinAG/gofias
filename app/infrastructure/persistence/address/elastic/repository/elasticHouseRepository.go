package repository

import (
	"context"
	"encoding/json"
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/domain/address/repository"
	"github.com/GarinAG/gofias/infrastructure/persistence/address/elastic/dto"
	elasticHelper "github.com/GarinAG/gofias/infrastructure/persistence/elastic"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/GarinAG/gofias/util"
	"github.com/dustin/go-humanize"
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
		  },
		  "analysis": {
			"filter": {
			  "russian_stemmer": {
				"type": "stemmer",
				"name": "russian"
			  },
			  "edge_ngram": {
				"type": "edge_ngram",
				"min_gram": "2",
				"max_gram": "25",
				"token_chars": ["letter", "digit"]
			  }
			},
			"analyzer": {
			  "edge_ngram_analyzer": {
				"filter": ["lowercase", "russian_stemmer", "edge_ngram"],
				"tokenizer": "standard"
			  },
			  "keyword_analyzer": {
				"filter": ["lowercase", "russian_stemmer"],
				"tokenizer": "standard"
			  }
			}
		  }
		}
	  },
	  "mappings": {
		"dynamic": false,
		"properties": {
		  "house_id": {
			"type": "keyword"
		  },
		  "house_guid": {
			"type": "keyword"
		  },
		  "ao_guid": {
			"type": "keyword"
		  },
		  "build_num": {
			"type": "keyword"
		  },
		  "house_num": {
			"type": "keyword"
		  },
		  "house_full_num": {
			"type": "text",
			"analyzer": "edge_ngram_analyzer",
			"search_analyzer": "keyword_analyzer",
			"fields": {
			  "keyword": {
				"type": "keyword"
			  }
			}
		  },
		  "str_num": {
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
		  "bazis_update_date": {
			"type": "date"
		  },
		  "update_date": {
			"type": "date"
		  },
		  "cad_num": {
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
	logger        interfaces.LoggerInterface
	batchSize     int
	indexName     string
}

func NewElasticHouseRepository(elasticClient *elasticHelper.Client, logger interfaces.LoggerInterface, batchSize int, prefix string) repository.HouseRepositoryInterface {
	return &ElasticHouseRepository{
		elasticClient: elasticClient,
		logger:        logger,
		batchSize:     batchSize,
		indexName:     prefix + entity.HouseObject{}.TableName(),
	}
}

func (a *ElasticHouseRepository) GetIndexName() string {
	return a.indexName
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

func (a *ElasticHouseRepository) GetByAddressGuid(guid string) ([]*entity.HouseObject, error) {
	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewTermQuery("ao_guid", guid)).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	var items []*entity.HouseObject
	var item *dto.JsonHouseDto
	if len(res.Hits.Hits) > 0 {
		for _, el := range res.Hits.Hits {
			if err := json.Unmarshal(el.Source, &item); err != nil {
				return nil, err
			}
		}
		items = append(items, item.ToEntity())
	}

	return items, nil
}

func (a *ElasticHouseRepository) InsertUpdateCollection(channel <-chan interface{}, done <-chan bool, count chan<- int) {
	bulk := a.elasticClient.Client.Bulk().Index(a.indexName).Pipeline(housesPipelineId)
	ctx := context.Background()
	var total uint64
	begin := time.Now()
	step := 1

Loop:
	for {
		select {
		case d := <-channel:
			if d == nil {
				break Loop
			}
			total++
			//util.PrintProcess(begin, total, 0, "house")
			saveItem := dto.JsonHouseDto{}
			saveItem.GetFromEntity(d.(entity.HouseObject))
			bulk.Add(elastic.NewBulkIndexRequest().Id(saveItem.ID).Doc(saveItem))
			if bulk.NumberOfActions() >= a.batchSize {
				res, err := bulk.Do(ctx)
				if err != nil {
					a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Add houses bulk commit failed")
				}
				if res.Errors {
					a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("Add houses bulk commit failed")
				}
				if total%uint64(100000) == 0 {
					a.logger.WithFields(interfaces.LoggerFields{"step": step, "count": total}).Info("Add houses to index")
					step++
				}
			}
		case <-done:
			break Loop
		}
	}

	// Commit the final batch before exiting
	if bulk.NumberOfActions() > 0 {
		res, err := bulk.Do(ctx)
		if err != nil {
			a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Add houses bulk commit failed")
		}
		if res.Errors {
			a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("Add houses bulk commit failed")
		}
		util.PrintProcess(begin, total, 0, "house")
	}
	a.logger.WithFields(interfaces.LoggerFields{"step": step, "count": total}).Info("Add houses to index")
	a.logger.WithFields(interfaces.LoggerFields{"execTime": humanize.RelTime(begin, time.Now(), "", "")}).Info("House import execution time")
	a.Refresh()
	count <- int(total)
}

func (a *ElasticHouseRepository) CountAllData() (int64, error) {
	return a.elasticClient.CountAllData(a.GetIndexName())
}

func (a *ElasticHouseRepository) Refresh() {
	a.elasticClient.RefreshIndexes([]string{a.GetIndexName()})
}
