package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GarinAG/gofias/domain/address/entity"
	"github.com/GarinAG/gofias/domain/address/repository"
	"github.com/GarinAG/gofias/infrastructure/persistence/address/elastic/dto"
	elasticHelper "github.com/GarinAG/gofias/infrastructure/persistence/elastic"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/GarinAG/gofias/util"
	"github.com/olivere/elastic/v7"
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
	logger        interfaces.LoggerInterface
	config        interfaces.ConfigInterface
	elasticClient *elasticHelper.Client
	indexName     string
	jobs          chan dto.JsonAddressDto
	results       chan dto.JsonAddressDto
}

func NewElasticAddressRepository(elasticClient *elasticHelper.Client, config interfaces.ConfigInterface, logger interfaces.LoggerInterface) repository.AddressRepositoryInterface {
	return &ElasticAddressRepository{
		logger:        logger,
		config:        config,
		elasticClient: elasticClient,
		indexName:     config.GetString("project.prefix") + entity.AddressObject{}.TableName(),
		jobs:          make(chan dto.JsonAddressDto, 10),
		results:       make(chan dto.JsonAddressDto, 10),
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

func (a *ElasticAddressRepository) InsertUpdateCollection(channel chan interface{}, done chan bool, count chan int) error {
	bulk := a.elasticClient.Client.Bulk().Index(a.indexName).Pipeline(addrPipelineId)
	ctx := context.Background()
	begin := time.Now()
	var total uint64
	for d := range channel {
		total++
		saveItem := a.ConvertToDto(d.(entity.AddressObject))
		util.PrintProcess(begin, total, 0, "item")
		// Enqueue the document
		bulk.Add(elastic.NewBulkIndexRequest().Id(saveItem.ID).Doc(saveItem))
		if bulk.NumberOfActions() >= a.config.GetInt("batch.size") {
			// Commit
			res, err := bulk.Do(ctx)
			if err != nil {
				return err
			}
			if res.Errors {
				return errors.New("Add addresses bulk commit failed")
			}
		}

		select {
		default:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Commit the final batch before exiting
	if bulk.NumberOfActions() > 0 {
		res, err := bulk.Do(ctx)
		util.PrintProcess(begin, total, 0, "item")
		if err != nil {
			return err
		}
		if res.Errors {
			return errors.New("Add addresses bulk commit failed")
		}
	}
	count <- int(total)

	return nil
}

func (a *ElasticAddressRepository) Index(houseRepos repository.HouseRepositoryInterface, isFull bool, start time.Time) error {
	a.elasticClient.RefreshIndexes([]string{a.GetIndexName(), houseRepos.GetIndexName()})
	a.elasticClient.Client.CloseIndex(a.GetIndexName())
	a.elasticClient.Client.OpenIndex(a.GetIndexName())

	query := elastic.NewBoolQuery() /*.Filter(elastic.NewTermQuery("ao_level", "7"))*/
	if !isFull {
		a.logger.Info("Indexing...")
		query.Must(elastic.NewRangeQuery("bazis_update_date").Gte(start))
	} else {
		a.logger.Info("Full indexing...")
	}

	scrollService := a.elasticClient.Client.Scroll(a.GetIndexName()).Query(query).Sort("ao_level", true)
	scanRes, err := a.elasticClient.ScrollData(scrollService)
	if err != nil {
		a.logger.Error(err.Error())
	}
	addrUpdateCount := len(scanRes)
	addTotalCount, err := a.elasticClient.CountAllData(a.GetIndexName())
	houseTotalCount, err := a.elasticClient.CountAllData(houseRepos.GetIndexName())

	a.logger.Info(fmt.Sprintf("Address update count: %d", addrUpdateCount))
	a.logger.Info(fmt.Sprintf("Total address count: %d", addTotalCount))
	a.logger.Info(fmt.Sprintf("Total houses count: %d", houseTotalCount))

	if addrUpdateCount > 0 {
		go a.allocate(scanRes)
		done := make(chan bool)
		var total uint64
		go a.result(done, time.Now(), total)
		noOfWorkers := 10
		a.createWorkerPool(noOfWorkers, houseRepos)
		<-done
	}
	a.logger.Info("Index Finished")

	return nil
}

func (a *ElasticAddressRepository) createWorkerPool(noOfWorkers int, houseRepos repository.HouseRepositoryInterface) {
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go a.searchAddressWorker(&wg, houseRepos)
	}
	wg.Wait()
	close(a.results)
}

func (a *ElasticAddressRepository) allocate(scanRes []elastic.SearchHit) {
	for _, addressItem := range scanRes {
		var item dto.JsonAddressDto
		if err := json.Unmarshal(addressItem.Source, &item); err != nil {
			a.logger.Fatal(err.Error())
		}
		a.jobs <- item
	}
	close(a.jobs)
}

func (a *ElasticAddressRepository) result(done chan bool, begin time.Time, total uint64) {
	bulk := a.elasticClient.Client.Bulk().Index(a.GetIndexName()).Pipeline(addrPipelineId)
	ctx := context.Background()

	for d := range a.results {
		total++
		util.PrintProcess(begin, total, 0, "item")
		// Enqueue the document
		bulk.Add(elastic.NewBulkIndexRequest().Id(d.ID).Doc(d))
		if bulk.NumberOfActions() >= a.config.GetInt("batch.size") {
			// Commit
			res, err := bulk.Do(ctx)
			if err != nil {
				a.logger.Fatal(err.Error())
				os.Exit(1)
			}
			if res.Errors {
				a.logger.Fatal("Bulk commit failed")
				os.Exit(1)
			}
		}
	}

	// Commit the final batch before exiting
	if bulk.NumberOfActions() > 0 {
		_, err := bulk.Do(ctx)
		if err != nil {
			a.logger.Fatal(err.Error())
			os.Exit(1)
		}
		util.PrintProcess(begin, total, 0, "item")
	}
	fmt.Println("")

	done <- true
}

func (a *ElasticAddressRepository) searchAddressWorker(wg *sync.WaitGroup, houseRepos repository.HouseRepositoryInterface) {
	for address := range a.jobs {
		searchCity, err := a.elasticClient.Client.
			Search(a.GetIndexName()).
			Query(elastic.NewMatchQuery("ao_guid", address.ParentGuid)).
			Do(context.Background())
		if err != nil {
			a.logger.Fatal(err.Error())
			os.Exit(1)
		}

		if len(searchCity.Hits.Hits) == 0 {
			continue
		}

		var city dto.JsonAddressDto
		var district dto.JsonAddressDto
		var house dto.JsonHouseDto
		var houseList []dto.JsonHouseDto

		if err := json.Unmarshal(searchCity.Hits.Hits[0].Source, &city); err != nil {
			a.logger.Fatal(err.Error())
			os.Exit(1)
		}
		if city.ParentGuid == "" {
			district = city
		} else {
			searchDistrict, err := a.elasticClient.Client.
				Search(a.GetIndexName()).
				Query(elastic.NewMatchQuery("ao_guid", city.ParentGuid)).
				Do(context.Background())
			if err != nil {
				a.logger.Fatal(err.Error())
				os.Exit(1)
			}

			if len(searchDistrict.Hits.Hits) == 0 {
				continue
			}

			if err := json.Unmarshal(searchDistrict.Hits.Hits[0].Source, &district); err != nil {
				a.logger.Fatal(err.Error())
				os.Exit(1)
			}
		}

		searchHouses, err := a.elasticClient.Client.
			Search(houseRepos.GetIndexName()).
			Query(elastic.NewMatchQuery("ao_guid", address.AoGuid)).
			Do(context.Background())
		if err != nil {
			a.logger.Fatal(err.Error())
			os.Exit(1)
		}

		for _, houseData := range searchHouses.Hits.Hits {
			if err := json.Unmarshal(houseData.Source, &house); err != nil {
				a.logger.Fatal(err.Error())
				os.Exit(1)
			}
			houseList = append(houseList, house)
		}

		postalCode := address.PostalCode
		if postalCode != "" {
			postalCode += ", "
		}

		address.StreetType = strings.TrimSpace(address.ShortName)
		address.Street = strings.TrimSpace(address.OffName)
		address.Settlement = strings.TrimSpace(city.OffName)
		address.SettlementType = strings.TrimSpace(city.ShortName)
		address.District = strings.TrimSpace(district.OffName)
		address.DistrictType = strings.TrimSpace(district.ShortName)
		address.StreetAddressSuggest = strings.ToLower(address.District +
			" " + address.Settlement +
			" " + address.Street)
		address.FullAddress = postalCode +
			district.ShortName + " " + district.OffName + ", " +
			city.ShortName + " " + city.OffName + ", " +
			address.ShortName + " " + address.OffName
		address.Houses = houseList

		a.results <- address
	}

	wg.Done()
}

func (a *ElasticAddressRepository) ConvertToEntity(item dto.JsonAddressDto) entity.AddressObject {
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

func (a *ElasticAddressRepository) ConvertToDto(item entity.AddressObject) dto.JsonAddressDto {
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
