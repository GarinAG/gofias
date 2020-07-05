package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/olivere/elastic/v7"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"os"
	"sync/atomic"
	"time"
)

type HouseItem struct {
	HouseId    string `xml:"HOUSEID,attr"`
	AoGuid     string `xml:"AOGUID,attr"`
	RegionCode string `xml:"REGIONCODE,attr"`
	PostalCode string `xml:"POSTALCODE,attr"`
	Okato      string `xml:"OKATO,attr"`
	Oktmo      string `xml:"OKTMO,attr"`
	IfNsFl     string `xml:"IFNSFL,attr"`
	IfNsUl     string `xml:"IFNSUL,attr"`
	TerrIfNsFl string `xml:"TERRIFNSFL,attr"`
	TerrIfNsUl string `xml:"TERRIFNSUL,attr"`
	NormDoc    string `xml:"NORMDOC,attr"`
	StartDate  string `xml:"STARTDATE,attr"`
	EndDate    string `xml:"ENDDATE,attr"`
	UpdateDate string `xml:"UPDATEDATE,attr"`
	DivType    string `xml:"DIVTYPE,attr"`
	HouseNum   string `xml:"HOUSENUM,attr"`
	BuildNum   string `xml:"BUILDNUM,attr"`
	StructNum  string `xml:"STRUCNUM,attr"`
	Counter    string `xml:"COUNTER,attr"`
	CadNum     string `xml:"CADNUM,attr"`
}

func importHouse(filePath string) uint64 {
	log.Printf("Start import file: %s", filePath)
	xmlStream, err := os.Open(filePath)
	if err != nil {
		log.Printf("Failed to open XML file: %s", filePath)
		return 0
	}
	defer xmlStream.Close()

	// Setup a group of goroutines from the excellent errgroup package
	g, ctx := errgroup.WithContext(context.TODO())
	docsc := make(chan HouseItemElastic)
	begin := time.Now()
	decoder := xml.NewDecoder(xmlStream)
	var elasticItem HouseItemElastic

	// Goroutine to create documents
	g.Go(func() error {
		defer close(docsc)
		for {
			// Read tokens from the XML document in a stream.
			t, err := decoder.Token()

			// If we are at the end of the file, we are done
			if err == io.EOF {
				log.Println("")
				break
			} else if err != nil {
				log.Fatalf("Error decoding token: %s", err)
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
					var item HouseItem

					// We decode the element into our data model...
					if err = decoder.DecodeElement(&item, &se); err != nil {
						log.Fatalf("Error decoding item: %s", err)
					}

					elasticItem = getHouseElasticStructFromXml(item)

					// Send over to 2nd goroutine, or cancel
					select {
					case docsc <- elasticItem:
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
		bulk := elasticClient.Bulk().Index(getPrefixIndexName(houseIndexName)).Pipeline(housesPipeline)
		for d := range docsc {
			if *status {
				// Simple progress
				current := atomic.AddUint64(&total, 1)
				dur := time.Since(begin).Seconds()
				sec := int(dur)
				pps := int64(float64(current) / dur)
				fmt.Printf("%10d | %6d req/s | %02d:%02d\r", current, pps, sec/60, sec%60)
			}

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

			if *status {
				// Final results
				dur := time.Since(begin).Seconds()
				sec := int(dur)
				pps := int64(float64(total) / dur)
				fmt.Printf("%10d | %6d req/s | %02d:%02d\n", total, pps, sec/60, sec%60)
			}
		}
		return nil
	})

	// Wait until all goroutines are finished
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}

	log.Println("Import Finished")

	return total
}
