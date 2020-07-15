package repository

import (
	"context"
	"encoding/json"
	"fmt"
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
	"strings"
	"sync"
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
          "max_ngram_diff": "18",
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
			  "ru_stop": {
				"type": "stop",
				"stopwords": "_russian_"
			  },
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
			  },
			  "index_analyzer": {
				"type": "custom",
				"tokenizer": "ngram-tokenizer",
				"filter": ["lowercase", "ru_stop", "trim"]
			  },
			  "search_analyzer": {
				"type": "custom",
				"tokenizer": "standard",
				"filter": ["lowercase", "ru_stop", "trim"]
			  }
			},
			"tokenizer": {
			  "ngram-tokenizer": {
				"type": "ngram",
				"min_gram": 2,
				"max_gram": 20,
                "token_chars": [
                  "letter",
                  "digit",
                  "punctuation",
                  "symbol"
                ]
			  }
			}
		  }
		}
	  },
	  "mappings": {
		"dynamic": false,
		"properties": {
		  "address_suggest": {
			"type": "completion",
			"analyzer": "autocomplete",
			"search_analyzer": "stop_analyzer"
		  },
		  "full_address": {
			"type": "text",
			"analyzer": "index_analyzer",
			"search_analyzer": "search_analyzer",
			"fields": {
			  "keyword": {
				"type": "keyword"
			  }
			}
		  },
		  "formal_name": {
			"type": "text",
			"analyzer": "index_analyzer",
			"search_analyzer": "search_analyzer",
			"fields": {
			  "keyword": {
				"type": "keyword"
			  }
			}
		  },
		  "formal_name_full": {
			"type": "keyword"
		  },
		  "ao_id": {
			"type": "keyword"
		  },
		  "ao_guid": {
			"type": "keyword"
		  },
		  "parent_guid": {
			"type": "keyword"
		  },
		  "ao_level": {
			"type": "integer"
		  },
		  "code": {
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
		  "act_status": {
			"type": "integer"
		  },
		  "live_status": {
			"type": "integer"
		  },
		  "postal_code": {
			"type": "keyword"
		  },
		  "region_code": {
			"type": "keyword"
		  },
		  "district": {
			"type": "keyword"
		  },
		  "district_type": {
			"type": "keyword"
		  },
		  "district_full": {
			"type": "keyword"
		  },
		  "settlement": {
			"type": "keyword"
		  },
		  "settlement_type": {
			"type": "keyword"
		  },
		  "settlement_full": {
			"type": "keyword"
		  },
		  "street": {
			"type": "keyword"
		  },
		  "street_type": {
			"type": "keyword"
		  },
		  "street_full": {
			"type": "keyword"
		  },
		  "okato": {
			"type": "keyword"
		  },
		  "oktmo": {
			"type": "keyword"
		  },
		  "start_date": {
			"type": "date"
		  },
		  "end_date": {
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
			  "house_id": {
				"type": "keyword"
			  },
			  "house_guid": {
				"type": "keyword"
			  },
			  "build_num": {
				"type": "keyword"
			  },
			  "house_num": {
				"type": "text",
				"analyzer": "index_analyzer",
				"search_analyzer": "search_analyzer",
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
	logger        interfaces.LoggerInterface
	batchSize     int
	elasticClient *elasticHelper.Client
	indexName     string
	jobs          chan dto.JsonAddressDto
	results       chan dto.JsonAddressDto
}

func NewElasticAddressRepository(elasticClient *elasticHelper.Client, logger interfaces.LoggerInterface, batchSize int, prefix string) repository.AddressRepositoryInterface {
	return &ElasticAddressRepository{
		logger:        logger,
		elasticClient: elasticClient,
		batchSize:     batchSize,
		indexName:     prefix + entity.AddressObject{}.TableName(),
	}
}

func (a *ElasticAddressRepository) Init() error {
	err := a.elasticClient.CreateIndex(a.indexName, addrIndexSettings)
	if err != nil {
		return err
	}

	return a.elasticClient.CreatePreprocessor(addrPipelineId, addrDropPipeline)
}

func (a *ElasticAddressRepository) GetIndexName() string {
	return a.indexName
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

	var item *dto.JsonAddressDto
	if len(res.Hits.Hits) > 0 {
		if err := json.Unmarshal(res.Hits.Hits[0].Source, &item); err != nil {
			return nil, err
		}

		return item.ToEntity(), nil
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

	var item *dto.JsonAddressDto
	if len(res.Hits.Hits) > 0 {
		if err := json.Unmarshal(res.Hits.Hits[0].Source, &item); err != nil {
			return nil, err
		}

		return item.ToEntity(), nil
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

	var item *dto.JsonAddressDto
	if len(res.Hits.Hits) > 0 {
		if err := json.Unmarshal(res.Hits.Hits[0].Source, &item); err != nil {
			return nil, err
		}

		return item.ToEntity(), nil
	}

	return nil, nil
}

func (a *ElasticAddressRepository) CountAllData() (int64, error) {
	return a.elasticClient.CountAllData(a.GetIndexName())
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
	var item *dto.JsonAddressDto
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
	var item *dto.JsonAddressDto
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

func (a *ElasticAddressRepository) GetDataByQuery(query elastic.Query) ([]elastic.SearchHit, error) {
	scrollService := a.elasticClient.Client.Scroll(a.GetIndexName()).Query(query).Sort("ao_level", true)
	return a.elasticClient.ScrollData(scrollService)
}

func (a *ElasticAddressRepository) GetBulkService() *elastic.BulkService {
	return a.elasticClient.Client.Bulk().Index(a.GetIndexName()).Pipeline(addrPipelineId)
}

func (a *ElasticAddressRepository) InsertUpdateCollection(channel <-chan interface{}, done <-chan bool, count chan<- int) {
	bulk := a.elasticClient.Client.Bulk().Index(a.indexName).Pipeline(addrPipelineId)
	ctx := context.Background()
	begin := time.Now()
	var total uint64
	step := 1

Loop:
	for {
		select {
		case d := <-channel:
			if d == nil {
				break Loop
			}
			total++
			saveItem := dto.JsonAddressDto{}
			saveItem.GetFromEntity(d.(entity.AddressObject))
			util.PrintProcess(begin, total, 0, "address")
			// Enqueue the document
			bulk.Add(elastic.NewBulkIndexRequest().Id(saveItem.ID).Doc(saveItem))
			if bulk.NumberOfActions() >= a.batchSize {
				// Commit
				res, err := bulk.Do(ctx)
				if err != nil {
					a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Add addresses bulk commit failed")
				}
				if res.Errors {
					a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("Add addresses bulk commit failed")
				}
				if total%uint64(a.batchSize*10) == 0 {
					a.logger.WithFields(interfaces.LoggerFields{"step": step, "count": total}).Info("Add addresses to index")
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
		util.PrintProcess(begin, total, 0, "address")
		if err != nil {
			a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Add addresses bulk commit failed")
		}
		if res.Errors {
			a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("Add addresses bulk commit failed")
		}
	}
	a.logger.WithFields(interfaces.LoggerFields{"step": step, "count": total}).Info("Add addresses to index")
	a.logger.WithFields(interfaces.LoggerFields{"execTime": humanize.RelTime(begin, time.Now(), "", "")}).Info("Address index execution time")
	a.Refresh()

	count <- int(total)
}

func (a *ElasticAddressRepository) Refresh() {
	a.elasticClient.RefreshIndexes([]string{a.GetIndexName()})
}

func (a *ElasticAddressRepository) ReopenIndex() {
	a.elasticClient.Client.CloseIndex(a.GetIndexName())
	a.elasticClient.Client.OpenIndex(a.GetIndexName())
}

func (a *ElasticAddressRepository) Index(isFull bool, start time.Time, housesCount int64, GetHousesByGuid repository.GetHousesByGuid) error {
	noOfWorkers := 10
	a.jobs = make(chan dto.JsonAddressDto, noOfWorkers)
	a.results = make(chan dto.JsonAddressDto, noOfWorkers)
	time.Sleep(1 * time.Second)
	a.Refresh()
	a.ReopenIndex()

	queries := []elastic.Query{elastic.NewRangeQuery("ao_level").Gt(1)}
	if !isFull {
		a.logger.Info("Indexing...")
		queries = append(queries, elastic.NewRangeQuery("bazis_update_date").Gte(start))
	} else {
		a.logger.Info("Full indexing...")
	}
	query := elastic.NewBoolQuery().Must(queries...)

	addTotalCount, err := a.CountAllData()
	if err != nil {
		a.logger.Error(err.Error())
	}

	a.logger.WithFields(interfaces.LoggerFields{"count": addTotalCount}).Info("Total address count")
	a.logger.WithFields(interfaces.LoggerFields{"count": housesCount}).Info("Total houses count")

	go a.allocate(query)
	done := make(chan bool)
	var total uint64
	go a.result(done, time.Now(), total)
	a.createWorkerPool(noOfWorkers, GetHousesByGuid)
	<-done
	a.Refresh()
	a.logger.Info("Index Finished")

	return nil
}

func (a *ElasticAddressRepository) allocate(query elastic.Query) {
	scrollService := a.elasticClient.Client.Scroll(a.GetIndexName()).
		Query(query).
		Sort("ao_level", true).
		Size(a.batchSize)

	ctx := context.Background()
	scrollService.Scroll("1h")
	count := 0
	var wg sync.WaitGroup

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
		count += len(res.Hits.Hits)
		wg.Add(1)
		go a.addJobs(res.Hits.Hits, &wg)
	}

	wg.Wait()
	a.logger.WithFields(interfaces.LoggerFields{"count": count}).Info("Address update count")

	close(a.jobs)
}

func (a *ElasticAddressRepository) addJobs(hits []*elastic.SearchHit, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, hit := range hits {
		var item dto.JsonAddressDto
		if err := json.Unmarshal(hit.Source, &item); err != nil {
			a.logger.Fatal(err.Error())
		}
		a.jobs <- item
	}
}

func (a *ElasticAddressRepository) createWorkerPool(noOfWorkers int, GetHousesByGuid repository.GetHousesByGuid) {
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go a.searchAddressWorker(&wg, GetHousesByGuid)
	}
	wg.Wait()
	close(a.results)
}

func (a *ElasticAddressRepository) searchAddressWorker(wg *sync.WaitGroup, GetHousesByGuid repository.GetHousesByGuid) {
	for address := range a.jobs {
		var houseList []dto.JsonHouseDto
		dtoItem := dto.JsonAddressDto{}
		city := dto.JsonAddressDto{}
		district := dto.JsonAddressDto{}
		guid := address.ParentGuid
		address.FullName = util.PrepareFullName(address.ShortName, address.OffName)
		address.FullAddress = address.FullName
		address.AddressSuggest = strings.TrimSpace(address.OffName)

		for guid != "" {
			search, _ := a.GetByGuid(guid)
			if search != nil {
				dtoItem.GetFromEntity(*search)
				guid = dtoItem.ParentGuid
				address.FullAddress = util.PrepareFullName(dtoItem.ShortName, dtoItem.OffName) + ", " + address.FullAddress
				address.AddressSuggest = strings.TrimSpace(dtoItem.OffName) + " " + address.AddressSuggest

				if dtoItem.AoLevel >= 4 {
					city = dtoItem
				}
				if dtoItem.AoLevel < 4 {
					district = dtoItem
				}
			} else {
				guid = ""
			}
		}

		if district.ID == "" && city.ID != "" {
			district = city
		}

		address.District = strings.TrimSpace(district.OffName)
		address.DistrictType = strings.TrimSpace(district.ShortName)
		address.Settlement = strings.TrimSpace(city.OffName)
		address.SettlementType = strings.TrimSpace(city.ShortName)

		if address.District != "" {
			address.DistrictFull = util.PrepareFullName(address.DistrictType, address.District)
		}
		if address.Settlement != "" {
			if address.DistrictFull != "" {
				address.SettlementFull = address.DistrictFull + ", "
			}
			address.SettlementFull += util.PrepareFullName(address.SettlementType, address.Settlement)
		}

		switch address.AoLevel {
		case 7:
			address.StreetType = strings.TrimSpace(address.ShortName)
			address.Street = strings.TrimSpace(address.OffName)
			searchHouses := GetHousesByGuid(address.AoGuid)
			if address.SettlementFull != "" {
				address.StreetFull = address.SettlementFull + ", "
			} else {
				if address.DistrictFull != "" {
					address.StreetFull = address.DistrictFull + ", "
				}
			}
			address.StreetFull += util.PrepareFullName(address.StreetType, address.Street)

			for _, houseData := range searchHouses {
				houseItem := dto.JsonHouseDto{}
				houseItem.GetFromEntity(*houseData)
				houseList = append(houseList, houseItem)
				address.Houses = houseList
			}
		}

		address.AddressSuggest = strings.ToLower(address.AddressSuggest)

		a.results <- address
	}

	wg.Done()
}

func (a *ElasticAddressRepository) result(done chan bool, begin time.Time, total uint64) {
	bulk := a.GetBulkService()
	ctx := context.Background()

	for d := range a.results {
		total++
		util.PrintProcess(begin, total, 0, "address")
		// Enqueue the document
		bulk.Add(elastic.NewBulkIndexRequest().Id(d.ID).Doc(d))
		if bulk.NumberOfActions() >= a.batchSize {
			// Commit
			res, err := bulk.Do(ctx)
			if err != nil {
				a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Index bulk commit failed")
				os.Exit(1)
			}
			if res.Errors {
				a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("Index bulk commit failed")
				os.Exit(1)
			}
		}
	}

	// Commit the final batch before exiting
	if bulk.NumberOfActions() > 0 {
		res, err := bulk.Do(ctx)
		if err != nil {
			a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Index bulk commit failed")
			os.Exit(1)
		}
		if res.Errors {
			a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("Index bulk commit failed")
			os.Exit(1)
		}
		util.PrintProcess(begin, total, 0, "address")
	}
	fmt.Println("")
	a.logger.WithFields(interfaces.LoggerFields{"count": total}).Info("Number of indexed addresses")
	a.logger.WithFields(interfaces.LoggerFields{"execTime": humanize.RelTime(begin, time.Now(), "", "")}).Info("Address index execution time")

	done <- true
}
