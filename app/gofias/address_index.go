package main

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
	"strings"
	"sync"
	"time"
)

var jobs = make(chan AddressItemElastic, 10)
var results = make(chan AddressItemElastic, 10)

func getAddressStructFromSearchHits(scanRes []elastic.SearchHit) []AddressItemElastic {
	var item AddressItemElastic
	var result []AddressItemElastic

	for _, res := range scanRes {
		if err := json.Unmarshal(res.Source, &item); err != nil {
			logFatal(err)
		}
		result = append(result, item)
	}

	return result
}

func CreateAddressIndex() {
	RefreshIndexes()
	elasticClient.CloseIndex(GetPrefixIndexName(addressIndexName))
	elasticClient.OpenIndex(GetPrefixIndexName(addressIndexName))

	query := elastic.NewBoolQuery().Filter(elastic.NewTermQuery("ao_level", "7"))
	if isUpdate {
		logPrintln("Indexing...")
		query.Must(elastic.NewTermQuery("bazis_update_date", versionDate))
	} else {
		logPrintln("Full indexing...")
	}

	scrollService := elasticClient.Scroll(GetPrefixIndexName(addressIndexName)).Query(query)
	scanRes := ScrollData(scrollService)
	addrUpdateCount := len(scanRes)

	logPrintf("Address update count: %d", addrUpdateCount)
	logPrintf("Total address count: %d", countAllData(addressIndexName))
	logPrintf("Total houses count: %d", countAllData(houseIndexName))

	if addrUpdateCount > 0 {
		go allocate(scanRes)
		done := make(chan bool)
		var total uint64
		go result(done, time.Now(), total)
		noOfWorkers := 10
		createWorkerPool(noOfWorkers)
		<-done
	}
	logPrintln("Index Finished")
}

func createWorkerPool(noOfWorkers int) {
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go searchAddressWorker(&wg)
	}
	wg.Wait()
	close(results)
}

func allocate(scanRes []elastic.SearchHit) {
	for _, addressItem := range scanRes {
		var item AddressItemElastic
		if err := json.Unmarshal(addressItem.Source, &item); err != nil {
			logFatal(err)
		}
		jobs <- item
	}
	close(jobs)
}

func result(done chan bool, begin time.Time, total uint64) {
	bulk := elasticClient.Bulk().Index(GetPrefixIndexName(addressIndexName)).Pipeline(addrPipeline)
	ctx := context.Background()

	for d := range results {
		total++
		PrintProcess(begin, total, 0, "item")
		// Enqueue the document
		bulk.Add(elastic.NewBulkIndexRequest().Id(d.ID).Doc(d))
		if bulk.NumberOfActions() >= *bulkSize {
			// Commit
			res, err := bulk.Do(ctx)
			if err != nil {
				logFatal(err)
			}
			if res.Errors {
				logFatal("Bulk commit failed")
			}
		}
	}

	// Commit the final batch before exiting
	if bulk.NumberOfActions() > 0 {
		_, err := bulk.Do(ctx)
		if err != nil {
			logFatal(err)
		}
		PrintProcess(begin, total, 0, "item")
	}
	fmtPrintln("")

	done <- true
}

func searchAddressWorker(wg *sync.WaitGroup) {
	for address := range jobs {
		searchCity, err := elasticClient.
			Search(GetPrefixIndexName(addressIndexName)).
			Query(elastic.NewMatchQuery("ao_guid", address.ParentGuid)).
			Do(context.Background())
		if err != nil {
			logFatal(err)
		}

		if len(searchCity.Hits.Hits) == 0 {
			continue
		}

		var city AddressItemElastic
		var district AddressItemElastic
		var house HouseItemElastic
		var houseList []HouseItemElastic

		if err := json.Unmarshal(searchCity.Hits.Hits[0].Source, &city); err != nil {
			logFatal(err)
		}
		if city.ParentGuid == "" {
			district = city
		} else {
			searchDistrict, err := elasticClient.
				Search(GetPrefixIndexName(addressIndexName)).
				Query(elastic.NewMatchQuery("ao_guid", city.ParentGuid)).
				Do(context.Background())
			if err != nil {
				logFatal(err)
			}

			if len(searchDistrict.Hits.Hits) == 0 {
				continue
			}

			if err := json.Unmarshal(searchDistrict.Hits.Hits[0].Source, &district); err != nil {
				logFatal(err)
			}
		}

		if IndexExists(houseIndexName) {
			searchHouses, err := elasticClient.
				Search(GetPrefixIndexName(houseIndexName)).
				Query(elastic.NewMatchQuery("ao_guid", address.AoGuid)).
				Do(context.Background())
			if err != nil {
				logFatal(err)
			}

			for _, houseData := range searchHouses.Hits.Hits {
				if err := json.Unmarshal(houseData.Source, &house); err != nil {
					logFatal(err)
				}
				houseList = append(houseList, house)
			}
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

		results <- address
	}

	wg.Done()
}
