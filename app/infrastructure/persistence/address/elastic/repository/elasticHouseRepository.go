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
	"io"
	"os"
	"sync"
	"time"
)

const (
	houseIndexSettings = `
	{
	  "settings": {
		"index": {
		  "number_of_shards": 1,
		  "number_of_replicas": 0,
		  "refresh_interval": "5s",
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
			  "ngram": {
				"type": "ngram",
				"min_gram": "1",
				"max_gram": "15"
			  },
			  "edge_ngram": {
                "type": "edge_ngram",
                "min_gram": "1",
                "max_gram": "50"
			  }
			},
			"analyzer": {
			  "ngram_analyzer": {
				"filter": ["lowercase", "ngram"],
				"tokenizer": "standard"
			  },
			  "edge_ngram_analyzer": {
				"filter": ["lowercase", "edge_ngram"],
				"tokenizer": "standard"
			  },
			  "keyword_analyzer": {
				"filter": ["lowercase"],
				"tokenizer": "standard"
			  }
			}
		  },
          "max_ngram_diff": 14
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
          "address_suggest": {
            "type": "text",
            "analyzer": "edge_ngram_analyzer",
            "search_analyzer": "keyword_analyzer"
          },
		  "house_full_num": {
			"type": "text",
			"analyzer": "ngram_analyzer",
			"search_analyzer": "keyword_analyzer",
			"fields": {
			  "keyword": {
				"type": "keyword"
			  }
			}
		  },
		  "full_address": {
			"type": "keyword"
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
	results       chan dto.JsonHouseDto
	noOfWorkers   int
}

func NewElasticHouseRepository(elasticClient *elasticHelper.Client, logger interfaces.LoggerInterface, batchSize int, prefix string, noOfWorkers int) repository.HouseRepositoryInterface {
	return &ElasticHouseRepository{
		elasticClient: elasticClient,
		logger:        logger,
		batchSize:     batchSize,
		indexName:     prefix + entity.HouseObject{}.TableName(),
		noOfWorkers:   noOfWorkers,
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

func (a *ElasticHouseRepository) scroll(scrollService *elastic.ScrollService) ([]*entity.HouseObject, error) {
	var items []*entity.HouseObject
	var item *dto.JsonHouseDto

	batch := a.batchSize
	if batch > 10000 {
		batch = 10000
	}
	scrollService.Size(batch)
	ctx := context.Background()
	scrollService.Scroll("1s")

	for {
		res, err := scrollService.Do(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			a.logger.Error(err.Error())
			break
		}
		if res == nil || len(res.Hits.Hits) == 0 {
			break
		}
		for _, el := range res.Hits.Hits {
			if err := json.Unmarshal(el.Source, &item); err != nil {
				return nil, err
			}
			items = append(items, item.ToEntity())
		}
	}

	err := scrollService.Clear(ctx)
	if err != nil {
		a.logger.Error(err.Error())
	}

	return items, nil
}

func (a *ElasticHouseRepository) GetByAddressGuid(guid string) ([]*entity.HouseObject, error) {
	scrollService := a.elasticClient.Client.Scroll(a.GetIndexName()).
		Query(elastic.NewTermQuery("ao_guid", guid)).
		Sort("house_full_num.keyword", true)

	return a.scroll(scrollService)
}

func (a *ElasticHouseRepository) GetLastUpdatedGuids(start time.Time) ([]string, error) {
	var guids []string

	scrollService := a.elasticClient.Client.Scroll(a.GetIndexName()).
		Query(elastic.NewRangeQuery("bazis_update_date").Gte(start.Format("2006-01-02") + "T00:00:00Z"))

	items, err := a.scroll(scrollService)

	if err != nil {
		return nil, err
	}
	for _, item := range items {
		guids = append(guids, item.AoGuid)
	}
	guids = util.UniqueStringSlice(guids)

	return guids, nil
}

func (a *ElasticHouseRepository) GetAddressByTerm(term string, size int64, from int64) ([]*entity.HouseObject, error) {
	if size == 0 {
		size = 100
	}

	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewBoolQuery().Must(
			elastic.NewMatchQuery("address_suggest", term).Operator("and"))).
		From(int(from)).
		Size(int(size)).
		Sort("full_address", true).
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
			items = append(items, item.ToEntity())
		}
	}

	return items, nil
}

func (a *ElasticHouseRepository) InsertUpdateCollection(channel <-chan interface{}, done <-chan bool, count chan<- int, isFull bool) {
	bulk := a.elasticClient.Client.Bulk().Index(a.indexName)
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
			saveItem := dto.JsonHouseDto{}
			saveItem.GetFromEntity(d.(entity.HouseObject))
			if saveItem.IsActive() {
				bulk.Add(elastic.NewBulkIndexRequest().Id(saveItem.ID).Doc(saveItem))
			} else {
				bulk.Add(elastic.NewBulkDeleteRequest().Id(saveItem.ID))
			}

			if bulk.NumberOfActions() >= a.batchSize {
				res, err := bulk.Do(ctx)
				if err != nil {
					a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Add houses bulk commit failed")
				}
				if res.Errors {
					a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("Add houses bulk commit failed")
				}
				if total%uint64(100000) == 0 && !util.CanPrintProcess {
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
	}
	if !util.CanPrintProcess {
		a.logger.WithFields(interfaces.LoggerFields{"step": step, "count": total}).Info("Add houses to index")
	}
	a.logger.WithFields(interfaces.LoggerFields{"count": total, "execTime": humanize.RelTime(begin, time.Now(), "", "")}).Info("House import execution time")
	a.Refresh()
	count <- int(total)
}

func (a *ElasticHouseRepository) CountAllData(query interface{}) (int64, error) {
	if query == nil {
		query = elastic.NewBoolQuery()
	}
	return a.elasticClient.CountAllData(a.GetIndexName(), query.(elastic.Query))
}

func (a *ElasticHouseRepository) Refresh() {
	a.elasticClient.RefreshIndexes([]string{a.GetIndexName()})
}

func (a *ElasticHouseRepository) GetBulkService() *elastic.BulkService {
	return a.elasticClient.Client.Bulk().Index(a.GetIndexName())
}

func (a *ElasticHouseRepository) Index(indexChan <-chan entity.IndexObject) error {
	a.results = make(chan dto.JsonHouseDto, a.noOfWorkers)
	a.Refresh()

	done := make(chan bool)
	go a.saveIndexItems(done)
	a.createWorkerPool(a.noOfWorkers, indexChan)
	<-done
	a.Refresh()

	return nil
}

func (a *ElasticHouseRepository) createWorkerPool(noOfWorkers int, indexChan <-chan entity.IndexObject) {
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go a.prepareItemsBeforeSave(&wg, indexChan)
	}
	wg.Wait()
	close(a.results)
}

func (a *ElasticHouseRepository) prepareItemsBeforeSave(wg *sync.WaitGroup, indexChan <-chan entity.IndexObject) {
	for d := range indexChan {
		houses, err := a.GetByAddressGuid(d.AoGuid)
		if err != nil {
			a.logger.WithFields(interfaces.LoggerFields{"error": err, "ao_guid": d.AoGuid}).Fatal("Get houses failed")
			continue
		}
		for _, house := range houses {
			saveItem := dto.JsonHouseDto{}
			saveItem.GetFromEntity(*house)
			saveItem.UpdateBazisDate()

			suggest := "дом д. " + saveItem.HouseNum
			if saveItem.StructNum != "" {
				suggest += ", строение стр. " + saveItem.StructNum
			}
			if saveItem.BuildNum != "" {
				suggest += ", корпус кор. " + saveItem.BuildNum
			}

			saveItem.AddressSuggest = d.AddressSuggest + ", " + suggest
			saveItem.FullAddress = d.FullAddress + ", " + saveItem.HouseFullNum

			a.results <- saveItem
		}
	}

	wg.Done()
}

func (a *ElasticHouseRepository) saveIndexItems(done chan bool) {
	bulk := a.GetBulkService()
	ctx := context.Background()
	begin := time.Now()

	for d := range a.results {
		// Enqueue the document
		bulk.Add(elastic.NewBulkIndexRequest().Id(d.ID).Doc(d))
		if bulk.NumberOfActions() >= a.batchSize {
			// Commit
			res, err := bulk.Do(ctx)
			if err != nil {
				a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("House index bulk commit failed")
				os.Exit(1)
			}
			if res.Errors {
				a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("House index bulk commit failed")
				os.Exit(1)
			}
		}
	}

	// Commit the final batch before exiting
	if bulk.NumberOfActions() > 0 {
		res, err := bulk.Do(ctx)
		if err != nil {
			a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("House index bulk commit failed")
			os.Exit(1)
		}
		if res.Errors {
			a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("House index bulk commit failed")
			os.Exit(1)
		}
	}
	a.logger.WithFields(interfaces.LoggerFields{"execTime": humanize.RelTime(begin, time.Now(), "", "")}).Info("House index execution time")
	done <- true
}
