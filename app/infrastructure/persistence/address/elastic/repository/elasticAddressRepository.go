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
              "edge_ngram": {
                "type": "edge_ngram",
                "min_gram": "1",
                "max_gram": "40"
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
          "address_suggest": {
            "type": "text",
            "analyzer": "edge_ngram_analyzer",
            "search_analyzer": "keyword_analyzer"
          },
          "full_address": {
            "type": "text",
            "analyzer": "edge_ngram_analyzer",
            "search_analyzer": "keyword_analyzer",
            "fields": {
			  "keyword": {
				"type": "keyword"
			  }
			}
          },
          "formal_name": {
            "type": "keyword"
          },
          "full_name": {
            "type": "text",
            "analyzer": "edge_ngram_analyzer",
            "search_analyzer": "keyword_analyzer",
            "fields": {
			  "keyword": {
				"type": "keyword"
			  }
			}
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
              "house_full_num": {
                "type": "keyword"
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
		  "if": "ctx.curr_status != 0"
		}
	  }, {
		"drop": {
		  "if": "ctx.act_status != 1"
		}
	  }, {
		"drop": {
		  "if": "ctx.live_status != 1"
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
			Must(elastic.NewMatchQuery("full_name", term))).
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

func (a *ElasticAddressRepository) CountAllData(query interface{}) (int64, error) {
	if query == nil {
		query = elastic.NewBoolQuery()
	}
	return a.elasticClient.CountAllData(a.GetIndexName(), query.(elastic.Query))
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
			items = append(items, item.ToEntity())
		}
	}

	return items, nil
}

func (a *ElasticAddressRepository) GetCitiesByTerm(term string, size int64, from int64) ([]*entity.AddressObject, error) {
	if size == 0 {
		size = 100
	}

	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewBoolQuery().Must(
			elastic.NewMultiMatchQuery(term, "full_address").Operator("and")).
			Filter(elastic.NewTermsQuery("ao_level", 1, 4))).
		From(int(from)).
		Size(int(size)).
		Sort("ao_level", true).
		Sort("full_address.keyword", true).
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
			items = append(items, item.ToEntity())
		}
	}

	return items, nil
}

func (a *ElasticAddressRepository) GetAddressByTerm(term string, size int64, from int64) ([]*entity.AddressObject, error) {
	if size == 0 {
		size = 100
	}

	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewBoolQuery().Must(
			elastic.NewMatchQuery("full_address", term).Operator("and"))).
		From(int(from)).
		Size(int(size)).
		Sort("ao_level", true).
		Sort("full_address.keyword", true).
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
			items = append(items, item.ToEntity())
		}
	}

	return items, nil
}

func (a *ElasticAddressRepository) GetAddressByPostal(term string, size int64, from int64) ([]*entity.AddressObject, error) {
	if size == 0 {
		size = 100
	}
	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewBoolQuery().Filter(
			elastic.NewTermQuery("postal_code", term))).
		From(int(from)).
		Size(int(size)).
		Sort("ao_level", true).
		Sort("full_address.keyword", true).
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
			items = append(items, item.ToEntity())
		}
	}

	return items, nil
}

func (a *ElasticAddressRepository) GetDataByQuery(query elastic.Query) ([]elastic.SearchHit, error) {
	scrollService := a.elasticClient.Client.Scroll(a.GetIndexName()).Query(query).Sort("ao_level", true)
	return a.elasticClient.ScrollData(scrollService)
}

func (a *ElasticAddressRepository) GetBulkService() *elastic.BulkService {
	return a.elasticClient.Client.Bulk().Index(a.GetIndexName())
}

func (a *ElasticAddressRepository) InsertUpdateCollection(channel <-chan interface{}, done <-chan bool, count chan<- int, isFull bool) {
	bulk := a.GetBulkService()
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
			// Enqueue the document
			if saveItem.IsActive() {
				bulk.Add(elastic.NewBulkIndexRequest().Id(saveItem.ID).Doc(saveItem))
			} else {
				bulk.Add(elastic.NewBulkDeleteRequest().Id(saveItem.ID))
			}

			if bulk.NumberOfActions() >= a.batchSize {
				// Commit
				res, err := bulk.Do(ctx)
				if err != nil {
					a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Add addresses bulk commit failed")
				}
				if res.Errors {
					a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("Add addresses bulk commit failed")
				}
				if total%uint64(100000) == 0 && !util.CanPrintProcess {
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
		if err != nil {
			a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Add addresses bulk commit failed")
		}
		if res.Errors {
			a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("Add addresses bulk commit failed")
		}
	}
	if !util.CanPrintProcess {
		a.logger.WithFields(interfaces.LoggerFields{"step": step, "count": total}).Info("Add addresses to index")
	}
	a.logger.WithFields(interfaces.LoggerFields{"count": total, "execTime": humanize.RelTime(begin, time.Now(), "", "")}).Info("Address index execution time")
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

func (a *ElasticAddressRepository) Index(isFull bool, start time.Time, guids []string, indexChan chan<- entity.IndexObject) error {
	noOfWorkers := 10
	a.jobs = make(chan dto.JsonAddressDto, noOfWorkers)
	a.results = make(chan dto.JsonAddressDto, noOfWorkers)
	a.Refresh()
	a.ReopenIndex()

	query := a.prepareIndexQuery(isFull, start, guids)
	queryCount := a.calculateIndexCount(query)

	go a.getIndexItems(query)
	done := make(chan bool)
	go a.saveIndexItems(done, time.Now(), queryCount, indexChan)
	a.createWorkerPool(noOfWorkers)
	<-done
	a.Refresh()

	return nil
}

func (a *ElasticAddressRepository) prepareIndexQuery(isFull bool, start time.Time, guids []string) elastic.Query {
	var query elastic.Query
	queries := []elastic.Query{elastic.NewRangeQuery("ao_level").Gt(1)}
	if !isFull {
		a.logger.Info("Indexing...")
		queries = append(queries, elastic.NewRangeQuery("bazis_update_date").Gte(start.Format("2006-01-02")+"T00:00:00Z"))
		if len(guids) > 0 {
			guidsInterface := util.ConvertStringSliceToInterface(guids)
			query = elastic.NewBoolQuery().Should(elastic.NewBoolQuery().Must(queries...), elastic.NewBoolQuery().Must(elastic.NewTermsQuery("ao_guid", guidsInterface...)))
		} else {
			query = elastic.NewBoolQuery().Must(queries...)
		}
	} else {
		a.logger.Info("Full indexing...")
		query = elastic.NewBoolQuery().Must(queries...)
	}

	return query
}

func (a *ElasticAddressRepository) calculateIndexCount(query elastic.Query) int64 {
	addTotalCount, err := a.CountAllData(nil)
	if err != nil {
		a.logger.Error(err.Error())
	}
	queryCount, err := a.CountAllData(query)
	if err != nil {
		a.logger.Error(err.Error())
	}

	a.logger.WithFields(interfaces.LoggerFields{"count": addTotalCount}).Info("Total address count")
	a.logger.WithFields(interfaces.LoggerFields{"count": queryCount}).Info("Number of indexed addresses")

	return queryCount
}

func (a *ElasticAddressRepository) getIndexItems(query elastic.Query) {
	batch := a.batchSize
	if batch > 10000 {
		batch = 10000
	}

	scrollService := a.elasticClient.Client.Scroll(a.GetIndexName()).
		Query(query).
		Sort("ao_level", true).
		Size(batch)

	ctx := context.Background()
	scrollService.Scroll("1s")
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

	err := scrollService.Clear(ctx)
	if err != nil {
		a.logger.Error(err.Error())
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

func (a *ElasticAddressRepository) createWorkerPool(noOfWorkers int) {
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go a.prepareItemsBeforeSave(&wg)
	}
	wg.Wait()
	close(a.results)
}

func (a *ElasticAddressRepository) prepareItemsBeforeSave(wg *sync.WaitGroup) {
	for address := range a.jobs {
		dtoItem := dto.JsonAddressDto{}
		city := dto.JsonAddressDto{}
		district := dto.JsonAddressDto{}
		guid := address.ParentGuid
		address.FullName = util.PrepareFullName(address.ShortName, address.FormalName)
		address.FullAddress = address.FullName
		address.AddressSuggest = strings.TrimSpace(address.FormalName)

		for guid != "" {
			search, _ := a.GetByGuid(guid)
			if search != nil {
				dtoItem.GetFromEntity(*search)
				guid = dtoItem.ParentGuid
				address.FullAddress = util.PrepareFullName(dtoItem.ShortName, dtoItem.FormalName) + ", " + address.FullAddress
				address.AddressSuggest = strings.TrimSpace(dtoItem.FormalName) + " " + address.AddressSuggest

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

		address.District = strings.TrimSpace(district.FormalName)
		address.DistrictType = strings.TrimSpace(district.ShortName)
		address.Settlement = strings.TrimSpace(city.FormalName)
		address.SettlementType = strings.TrimSpace(city.ShortName)

		if address.District != "" {
			address.DistrictFull = util.PrepareFullName(address.DistrictType, address.District)
		}
		if address.Settlement != "" {
			address.SettlementFull = ""
			if address.DistrictFull != "" {
				address.SettlementFull = address.DistrictFull + ", "
			}
			address.SettlementFull += util.PrepareFullName(address.SettlementType, address.Settlement)
		}

		switch address.AoLevel {
		case 7:
			address.StreetType = strings.TrimSpace(address.ShortName)
			address.Street = strings.TrimSpace(address.FormalName)
			address.StreetFull = ""
			if address.SettlementFull != "" {
				address.StreetFull = address.SettlementFull + ", "
			} else {
				if address.DistrictFull != "" {
					address.StreetFull = address.DistrictFull + ", "
				}
			}
			address.StreetFull += util.PrepareFullName(address.StreetType, address.Street)
		}

		address.AddressSuggest = strings.ToLower(address.AddressSuggest)

		a.results <- address
	}

	wg.Done()
}

func (a *ElasticAddressRepository) saveIndexItems(done chan bool, begin time.Time, total int64, indexChan chan<- entity.IndexObject) {
	bulk := a.GetBulkService()
	ctx := context.Background()
	bar := util.StartNewProgress(int(total))

	for d := range a.results {
		// Enqueue the document
		bulk.Add(elastic.NewBulkIndexRequest().Id(d.ID).Doc(d))
		if d.AoLevel == 7 {
			indexChan <- entity.IndexObject{
				AoGuid:      d.AoGuid,
				FullAddress: d.FullAddress,
			}
		}
		bar.Increment()
		if bulk.NumberOfActions() >= a.batchSize {
			// Commit
			res, err := bulk.Do(ctx)
			if err != nil {
				a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Address index bulk commit failed")
				os.Exit(1)
			}
			if res.Errors {
				a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("Address index bulk commit failed")
				os.Exit(1)
			}
		}
	}

	// Commit the final batch before exiting
	if bulk.NumberOfActions() > 0 {
		res, err := bulk.Do(ctx)
		if err != nil {
			a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Address index bulk commit failed")
			os.Exit(1)
		}
		if res.Errors {
			a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("Address index bulk commit failed")
			os.Exit(1)
		}
	}
	bar.Finish()
	a.logger.WithFields(interfaces.LoggerFields{"execTime": humanize.RelTime(begin, time.Now(), "", "")}).Info("Address index execution time")
	done <- true
}
