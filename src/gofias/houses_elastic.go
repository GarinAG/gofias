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
	ID              string `json:"_id" xml:"HOUSEID,attr"`
	AoGuid          string `json:"ao_guid" xml:"AOGUID,attr"`
	RegionCode      string `json:"region_code" xml:"REGIONCODE,attr"`
	PostalCode      string `json:"postal_code" xml:"POSTALCODE,attr"`
	Okato           string `json:"okato" xml:"OKATO,attr"`
	Oktmo           string `json:"oktmo" xml:"OKTMO,attr"`
	IfNsFl          string `json:"ifns_fl" xml:"IFNSFL,attr"`
	IfNsUl          string `json:"ifns_ul" xml:"IFNSUL,attr"`
	TerrIfNsFl      string `json:"terr_ifns_fl" xml:"TERRIFNSFL,attr"`
	TerrIfNsUl      string `json:"terr_ifns_ul" xml:"TERRIFNSUL,attr"`
	NormDoc         string `json:"norm_doc" xml:"NORMDOC,attr"`
	StartDate       string `json:"start_date" xml:"STARTDATE,attr"`
	EndDate         string `json:"end_date" xml:"ENDDATE,attr"`
	UpdateDate      string `json:"update_date" xml:"UPDATEDATE,attr"`
	DivType         string `json:"div_type" xml:"DIVTYPE,attr"`
	HouseNum        string `json:"house_num" xml:"HOUSENUM,attr"`
	BuildNum        string `json:"build_num" xml:"BUILDNUM,attr"`
	StructNum       string `json:"str_num" xml:"STRUCNUM,attr"`
	Counter         string `json:"counter" xml:"COUNTER,attr"`
	CadNum          string `json:"cad_num" xml:"CADNUM,attr"`
	BazisCreateDate string `json:"bazis_create_date"`
	BazisUpdateDate string `json:"bazis_update_date"`
	BazisFinishDate string `json:"bazis_finish_date"`
}

func (item *HouseItemElastic) SetBazisProps() {
	currentTime := time.Now().Format("2006-01-02") + dateTimeZone
	saveTime := currentTime
	if isUpdate {
		saveTime = versionDate
	}

	item.BazisCreateDate = currentTime
	item.BazisUpdateDate = saveTime
	item.BazisFinishDate = item.EndDate
}

func importHouse(filePath string) uint64 {
	logPrintf("Start import file: %s", filePath)
	xmlStream, err := os.Open(filePath)
	if err != nil {
		logPrintf("Failed to open XML file: %s", filePath)
		return 0
	}
	defer xmlStream.Close()

	// Setup a group of goroutines from the excellent errgroup package
	g, ctx := errgroup.WithContext(context.TODO())
	docsc := make(chan HouseItemElastic)
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
				case housesTag:
					var item HouseItemElastic

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
		bulk := elasticClient.Bulk().Index(GetPrefixIndexName(houseIndexName)).Pipeline(housesPipeline)
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
