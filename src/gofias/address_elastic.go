package main

import (
	"context"
	"encoding/xml"
	"errors"
	"github.com/olivere/elastic/v7"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
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
	ID                   string             `json:"_id" xml:"AOID,attr"`
	AoGuid               string             `json:"ao_guid" xml:"AOGUID,attr"`
	ParentGuid           string             `json:"parent_guid" xml:"PARENTGUID,attr"`
	FormalName           string             `json:"formal_name" xml:"FORMALNAME,attr"`
	OffName              string             `json:"off_name" xml:"OFFNAME,attr"`
	ShortName            string             `json:"short_name" xml:"SHORTNAME,attr"`
	AoLevel              string             `json:"ao_level" xml:"AOLEVEL,attr"`
	AreaCode             string             `json:"area_code" xml:"AREACODE,attr"`
	CityCode             string             `json:"city_code" xml:"CITYCODE,attr"`
	PlaceCode            string             `json:"place_code" xml:"PLACECODE,attr"`
	AutoCode             string             `json:"auto_code" xml:"AUTOCODE,attr"`
	PlanCode             string             `json:"plan_code" xml:"PLANCODE,attr"`
	StreetCode           string             `json:"street_code" xml:"STREETCODE,attr"`
	CTarCode             string             `json:"city_ar_code" xml:"CTARCODE,attr"`
	ExtrCode             string             `json:"extr_code" xml:"EXTRCODE,attr"`
	SextCode             string             `json:"sub_ext_code" xml:"SEXTCODE,attr"`
	Code                 string             `json:"code" xml:"CODE,attr"`
	RegionCode           string             `json:"region_code" xml:"REGIONCODE,attr"`
	PlainCode            string             `json:"plain_code" xml:"PLAINCODE,attr"`
	PostalCode           string             `json:"postal_code" xml:"POSTALCODE,attr"`
	Okato                string             `json:"okato" xml:"OKATO,attr"`
	Oktmo                string             `json:"oktmo" xml:"OKTMO,attr"`
	IfNsFl               string             `json:"ifns_fl" xml:"IFNSFL,attr"`
	IfNsUl               string             `json:"ifns_ul" xml:"IFNSUL,attr"`
	TerrIfNsFl           string             `json:"terr_ifns_fl" xml:"TERRIFNSFL,attr"`
	TerrIfNsUl           string             `json:"terr_ifns_ul" xml:"TERRIFNSUL,attr"`
	NormDoc              string             `json:"norm_doc" xml:"NORMDOC,attr"`
	ActStatus            string             `json:"act_status" xml:"ACTSTATUS,attr"`
	LiveStatus           string             `json:"live_status" xml:"LIVESTATUS,attr"`
	CurrStatus           string             `json:"curr_status" xml:"CURRSTATUS,attr"`
	OperStatus           string             `json:"oper_status" xml:"OPERSTATUS,attr"`
	StartDate            string             `json:"start_date" xml:"STARTDATE,attr"`
	EndDate              string             `json:"end_date" xml:"ENDDATE,attr"`
	UpdateDate           string             `json:"update_date" xml:"UPDATEDATE,attr"`
	StreetType           string             `json:"street_type"`
	Street               string             `json:"street"`
	Settlement           string             `json:"settlement"`
	SettlementType       string             `json:"settlement_type"`
	District             string             `json:"district"`
	DistrictType         string             `json:"district_type"`
	StreetAddressSuggest string             `json:"street_address_suggest"`
	FullAddress          string             `json:"full_address"`
	Houses               []HouseItemElastic `json:"houses"`
	BazisCreateDate      string             `json:"bazis_create_date"`
	BazisUpdateDate      string             `json:"bazis_update_date"`
	BazisFinishDate      string             `json:"bazis_finish_date"`
}

func (item *AddressItemElastic) SetBazisProps() {
	currentTime := time.Now().Format("2006-01-02") + dateTimeZone
	saveTime := currentTime
	if isUpdate {
		saveTime = versionDate
	}

	item.BazisCreateDate = currentTime
	item.BazisUpdateDate = saveTime
	item.BazisFinishDate = item.EndDate
}

func importAddress(filePath string) uint64 {
	logPrintf("Start import file: %s", filePath)
	xmlStream, err := os.Open(filePath)
	if err != nil {
		logPrintf("Failed to open XML file: %s", filePath)
		return 0
	}
	defer xmlStream.Close()

	// Setup a group of goroutines from the excellent errgroup package
	g, ctx := errgroup.WithContext(context.TODO())
	docsc := make(chan AddressItemElastic)
	begin := time.Now()
	decoder := xml.NewDecoder(xmlStream)

	// Goroutine to create documents
	g.Go(func() error {
		defer close(docsc)
		for {
			// Read tokens from the XML document in a stream.
			t, err := decoder.Token()

			// If we are at the end of the file, we are done
			if err == io.EOF {
				logPrintln("")
				break
			} else if err != nil {
				logFatalf("Error decoding token: %s", err)
			} else if t == nil {
				break
			}

			// Here, we inspect the token
			switch se := t.(type) {
			// We have the start of an element.
			// However, we have the complete token in t
			case xml.StartElement:
				switch se.Name.Local {
				// Found an item, so we process it
				case addrTag:
					var item AddressItemElastic

					// We decode the element into our data model...
					if err = decoder.DecodeElement(&item, &se); err != nil {
						logFatalf("Error decoding item: %s", err)
					}
					item.SetBazisProps()

					select {
					case docsc <- item:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}
		}

		return nil
	})

	var total uint64
	g.Go(func() error {
		bulk := elasticClient.Bulk().Index(GetPrefixIndexName(addressIndexName)).Pipeline(addrPipeline)
		for d := range docsc {
			total++
			PrintProcess(begin, total, 0, "item")
			// Enqueue the document
			bulk.Add(elastic.NewBulkIndexRequest().Id(d.ID).Doc(d))
			if bulk.NumberOfActions() >= *bulkSize {
				// Commit
				res, err := bulk.Do(ctx)
				if err != nil {
					return err
				}
				if res.Errors {
					return errors.New("Bulk commit failed\n")
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
			_, err = bulk.Do(ctx)
			if err != nil {
				return err
			}
			PrintProcess(begin, total, 0, "item")
		}
		return nil
	})

	// Wait until all goroutines are finished
	if err := g.Wait(); err != nil {
		logFatal(err)
	}

	fmtPrintln("")
	logPrintln("Import Finished")

	return total
}
