package service

import (
	"context"
	"fmt"
	"github.com/GarinAG/gofias/domain/address/repository"
	"github.com/GarinAG/gofias/domain/osm/entity"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/GarinAG/gofias/util"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"os"
	"strings"
)

type OsmService struct {
	addressRepo repository.AddressRepositoryInterface
	houseRepo   repository.HouseRepositoryInterface
	logger      interfaces.LoggerInterface
}

func NewOsmService(addressRepo repository.AddressRepositoryInterface, houseRepo repository.HouseRepositoryInterface, logger interfaces.LoggerInterface) *OsmService {
	return &OsmService{
		addressRepo: addressRepo,
		houseRepo:   houseRepo,
		logger:      logger,
	}
}

func (o *OsmService) Update() {
	o.parseFile("./central-fed-district-latest.osm.pbf")
}

func (o *OsmService) parseFile(filepath string) {
	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := osmpbf.New(context.Background(), f, 3)
	defer scanner.Close()

	addressChan := make(chan *entity.Node)
	housesChan := make(chan *entity.Node)
	done := make(chan bool)

	go o.scan(scanner, done, addressChan, housesChan)
	go o.updateAddresses(done, addressChan)

	<-done
	close(addressChan)
	close(housesChan)

	scanErr := scanner.Err()
	if scanErr != nil {
		panic(scanErr)
	}
}

func (o *OsmService) scan(scanner *osmpbf.Scanner, done chan<- bool, addressChan chan<- *entity.Node, housesChan chan<- *entity.Node) {
	tagList := "name"
	conditions := make(map[string][]string)
	for _, group := range strings.Split(tagList, ",") {
		conditions[group] = strings.Split(group, "+")
	}

	bar := util.StartNewProgress(-1)
	for scanner.Scan() {
		switch e := scanner.Object().(type) {
		case *osm.Node:
			//case *osm.Way:
			//case *osm.Relation:
			if e.Tags != nil {
				if o.hasTags(e.TagMap()) && o.containsValidTags(e.TagMap(), conditions) {
					bar.Increment()
					node := o.prepareItems(e)
					if node != nil {
						switch node.Type {
						case "place":
							addressChan <- node
						case "building":
							housesChan <- node
						}
					}
				}
			}
		}
	}

	bar.Finish()
	close(addressChan)
	close(housesChan)
	done <- true
}

func (o *OsmService) updateAddresses(done <-chan bool, addressChan <-chan *entity.Node) {
	address := make(chan interface{})
	addressCnt := make(chan int)
	go o.addressRepo.InsertUpdateCollection(address, done, addressCnt, true)

	for d := range addressChan {
		items, _ := o.addressRepo.GetAddressByTerm(d.Name, 1, 0)
		if len(items) > 0 {
			item := items[0]
			item.Location = fmt.Sprint(d.Lon, ",", d.Lat)
			if item.PostalCode == "" && d.PostalCode != "" {
				item.PostalCode = d.PostalCode
			}
			address <- *item
		}
	}
	close(address)
}

func (o *OsmService) updateHouses(done <-chan bool, housesChan <-chan *entity.Node) {
	houses := make(chan interface{})
	housesCnt := make(chan int)
	go o.houseRepo.InsertUpdateCollection(houses, done, housesCnt, true)

	for d := range housesChan {
		items, _ := o.addressRepo.GetAddressByTerm(d.Name, 1, 0)
		if len(items) > 0 {
			item := items[0]
			item.Location = fmt.Sprint(d.Lon, ",", d.Lat)
			if item.PostalCode == "" && d.PostalCode != "" {
				item.PostalCode = d.PostalCode
			}
			houses <- *item
		}
	}

	close(houses)
}

func (o *OsmService) prepareItems(e *osm.Node) *entity.Node {
	place := o.getTagByName(e.TagMap(), "place")

	postal := o.getTagByName(e.TagMap(), "addr:postcode")
	region := o.getTagByName(e.TagMap(), "addr:region")
	city := o.getTagByName(e.TagMap(), "addr:city")
	street := o.getTagByName(e.TagMap(), "addr:street")
	housenum := o.getTagByName(e.TagMap(), "addr:housenumber")
	name := o.getTagByName(e.TagMap(), "name")

	fullAddr := ""
	if region != "" && region != name {
		fullAddr = region
	}
	if city != "" && city != name {
		if fullAddr != "" {
			fullAddr += ", "
		}
		fullAddr += city
	}
	if street != "" && street != name {
		if fullAddr != "" {
			fullAddr += ", "
		}
		fullAddr += street
	}
	if housenum != "" && housenum != name {
		if fullAddr != "" {
			fullAddr += ", "
		}
		fullAddr += housenum
	} else if name != "" {
		if fullAddr != "" {
			fullAddr += ", "
		}
		fullAddr += name
	}

	if fullAddr != "" {
		if place != "" || (housenum != "" && street != "" && city != "") {
			fullAddr = util.Replace(fullAddr)

			node := entity.Node{
				Name:       fullAddr,
				Lat:        e.Lat,
				Lon:        e.Lon,
				PostalCode: postal,
			}

			if place != "" {
				node.Type = "place"
			} else {
				node.Type = "building"
			}

			return &node
		}
	}

	return nil
}

// check tags contain features from a groups of whitelists
func (o *OsmService) containsValidTags(tags map[string]string, group map[string][]string) bool {
	for _, list := range group {
		if o.matchTagsAgainstCompulsoryTagList(tags, list) {
			return true
		}
	}
	return false
}

// trim leading/trailing spaces from keys and values
func (o *OsmService) trimTags(tags map[string]string) map[string]string {
	trimmed := make(map[string]string)
	for k, v := range tags {
		trimmed[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return trimmed
}

// check if a tag list is empty or not
func (o *OsmService) hasTags(tags map[string]string) bool {
	n := len(tags)
	if n == 0 {
		return false
	}
	return true
}

// check tags contain features from a whitelist
func (o *OsmService) matchTagsAgainstCompulsoryTagList(tags map[string]string, tagList []string) bool {
	for _, name := range tagList {

		feature := strings.Split(name, "~")
		foundVal, foundKey := tags[feature[0]]

		// key check
		if !foundKey {
			return false
		}

		// value check
		if len(feature) > 1 {
			if foundVal != feature[1] {
				return false
			}
		}
	}

	return true
}

// check tags contain features from a whitelist
func (o *OsmService) getTagByName(tags map[string]string, name string) string {
	foundVal, foundKey := tags[name]
	// key check
	if !foundKey {
		return ""
	}

	return foundVal
}
