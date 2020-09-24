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

	size := 0
	var places []entity.Node
	var buildings []entity.Node
	tagList := "name"
	conditions := make(map[string][]string)
	for _, group := range strings.Split(tagList, ",") {
		conditions[group] = strings.Split(group, "+")
	}

	for scanner.Scan() {
		/*if size > 100 {
		    break
		}*/
		switch e := scanner.Object().(type) {
		case *osm.Node:
			//case *osm.Way:
			//case *osm.Relation:
			if e.Tags != nil {
				if o.hasTags(e.TagMap()) && o.containsValidTags(e.TagMap(), conditions) {
					place := o.getTagByName(e.TagMap(), "place")
					building := o.getTagByName(e.TagMap(), "building")

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
						if place != "" || (building != "" && street != "" && city != "") {
							fullAddr = util.Replace(fullAddr)
						} else {
							continue
						}

						node := entity.Node{
							Name:       fullAddr,
							Lat:        e.Lat,
							Lon:        e.Lon,
							PostalCode: postal,
						}

						if place != "" {
							places = append(places, node)
						} else {
							buildings = append(buildings, node)
						}
						size++
					}
				}
			}
		}
	}

	fmt.Printf("%+v\n", places)
	fmt.Printf("%+v\n", buildings)

	scanErr := scanner.Err()
	if scanErr != nil {
		panic(scanErr)
	}
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
