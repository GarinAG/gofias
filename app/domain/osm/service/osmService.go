package service

import (
	"context"
	"fmt"
	"github.com/GarinAG/gofias/domain/address/repository"
	"github.com/GarinAG/gofias/domain/directory/service"
	"github.com/GarinAG/gofias/domain/osm/entity"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/GarinAG/gofias/util"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"os"
	"strings"
)

// Сервис работы с OSM
type OsmService struct {
	addressRepo     repository.AddressRepositoryInterface // Репозиторий адресов
	houseRepo       repository.HouseRepositoryInterface   // Репозиторий домов
	logger          interfaces.LoggerInterface            // Логгер
	downloadService *service.DownloadService              // Сервис управления загрузкой файлов
	config          interfaces.ConfigInterface            // Конфигурация
}

// Инициализация сервиса
func NewOsmService(
	addressRepo repository.AddressRepositoryInterface,
	houseRepo repository.HouseRepositoryInterface,
	downloadService *service.DownloadService,
	logger interfaces.LoggerInterface,
	config interfaces.ConfigInterface,
) *OsmService {
	return &OsmService{
		addressRepo:     addressRepo,
		houseRepo:       houseRepo,
		logger:          logger,
		downloadService: downloadService,
		config:          config,
	}
}

// Обновляет данные местоположений
func (o *OsmService) Update() {
	// Скачивает файл с данными
	file, err := o.downloadService.DownloadFile(o.config.GetString("osm.url"), "russia.pbf")
	o.checkFatalError(err)
	if file != nil {
		// Разбирает файл с данными
		o.parseFile(file.Path)
	}
}

// Разбирает файл с данными
func (o *OsmService) parseFile(filepath string) {
	// Открывает файл с данными OSM
	f, err := os.Open(filepath)
	o.checkFatalError(err)
	defer f.Close()

	// Создает объект сканнера
	scanner := osmpbf.New(context.Background(), f, 3)
	defer scanner.Close()

	addressChan := make(chan *entity.Node)
	housesChan := make(chan *entity.Node)
	done := make(chan bool)
	// Проверяет наличие домов в БД
	housesCnt, _ := o.houseRepo.CountAllData(nil)
	if housesCnt == 0 {
		housesChan = nil
	}

	// Сканирует файл с данными OSM
	go o.scan(scanner, done, addressChan, housesChan)
	// Обновляет адреса
	go o.updateAddresses(done, addressChan)
	// При наличии домов разрешает обновление местоположений
	if housesChan != nil {
		// Обновляет дома
		go o.updateHouses(done, housesChan)
	}

	<-done
	scanErr := scanner.Err()
	if scanErr != nil {
		panic(scanErr)
	}
}

// Сканирует файл с данными OSM
func (o *OsmService) scan(scanner *osmpbf.Scanner, done chan<- bool, addressChan chan<- *entity.Node, housesChan chan<- *entity.Node) {
	// Создает список условий, проставляет обязательное условие - наличие названия
	tagList := "name"
	conditions := make(map[string][]string)
	for _, group := range strings.Split(tagList, ",") {
		conditions[group] = strings.Split(group, "+")
	}

	bar := util.StartNewProgress(-1)
	for scanner.Scan() {
		switch e := scanner.Object().(type) {
		// Элемент является объектом
		case *osm.Node:
			//case *osm.Way:
			//case *osm.Relation:
			if e.Tags != nil {
				// Проверяет условия
				if o.hasTags(e.TagMap()) && o.containsValidTags(e.TagMap(), conditions) {
					bar.Increment()
					// Проверяет и разбирает объект
					node := o.prepareItems(e)
					if node != nil {
						switch node.Type {
						case "place": // Объект является адресом
							addressChan <- node
						case "building": // Объект является домом
							if housesChan != nil {
								housesChan <- node
							}
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

// Обновляет адреса
func (o *OsmService) updateAddresses(done <-chan bool, addressChan <-chan *entity.Node) {
	address := make(chan interface{})
	addressCnt := make(chan int)
	// Сохраняет элементы в БД
	go o.addressRepo.InsertUpdateCollection(address, done, addressCnt, true)

	for d := range addressChan {
		// Ищет адреса в БД по названию
		items, _ := o.addressRepo.GetAddressByTerm(d.Name, 1, 0)
		if len(items) > 0 {
			item := items[0]
			location := fmt.Sprint(d.Lat, ",", d.Lon)
			// Сохраняет только адреса, у которых отличается местоположение или индекс с данными из OSM
			if item.Location != location || item.PostalCode != d.PostalCode {
				item.Location = location
				if item.PostalCode == "" && d.PostalCode != "" {
					item.PostalCode = d.PostalCode
				}
				address <- *item
			}
		}
	}
	close(address)
}

// Обновляет дома
func (o *OsmService) updateHouses(done <-chan bool, housesChan <-chan *entity.Node) {
	houses := make(chan interface{})
	housesCnt := make(chan int)
	// Сохраняет элементы в БД
	go o.houseRepo.InsertUpdateCollection(houses, done, housesCnt, true)

	for d := range housesChan {
		// Ищет дома в БД по адресу
		items, _ := o.houseRepo.GetAddressByTerm(d.Name, 1, 0)
		if len(items) > 0 {
			item := items[0]
			location := fmt.Sprint(d.Lat, ",", d.Lon)
			// Сохраняет только дома, у которых отличается местоположение или индекс с данными из OSM
			if item.Location != location || item.PostalCode != d.PostalCode {
				item.Location = location
				if item.PostalCode == "" && d.PostalCode != "" {
					item.PostalCode = d.PostalCode
				}
				houses <- *item
			}
		}
	}

	close(houses)
}

// Проверяет и разбирает объект
func (o *OsmService) prepareItems(e *osm.Node) *entity.Node {
	place := o.getTagByName(e.TagMap(), "place")
	official := strings.Split(o.getTagByName(e.TagMap(), "official_status"), ":")
	postal := o.getTagByName(e.TagMap(), "addr:postcode")
	region := o.getTagByName(e.TagMap(), "addr:region")
	district := o.getTagByName(e.TagMap(), "addr:district")
	city := o.getTagByName(e.TagMap(), "addr:city")
	street := o.getTagByName(e.TagMap(), "addr:street")
	housenum := o.getTagByName(e.TagMap(), "addr:housenumber")
	name := o.getTagByName(e.TagMap(), "name")
	if strings.Contains(district, "городской округ") {
		district = ""
	}

	// Проверяет наличие тегов у объекта
	if place != "" || (housenum != "" || street != "" || city != "") {
		if place != "" && place != "city" && place != "town" && region == "" && district == "" && city == "" {
			return nil
		}

		fullAddr := ""
		if region != "" && region != name {
			fullAddr = region
		}
		if district != "" && district != name {
			if fullAddr != "" {
				fullAddr += ", "
			}
			fullAddr += district
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
		} else if name != "" && name != city {
			if fullAddr != "" {
				fullAddr += ", "
			}
			if len(official) > 0 && len(official[len(official)-1]) > 0 {
				name = official[len(official)-1] + " " + name
			}

			fullAddr += name
		}

		replacedAddr := strings.TrimSpace(util.Replace(fullAddr))
		node := entity.Node{
			Name:       replacedAddr,
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

	return nil
}

// Проверяет наличие конкретных тегов у объекта по группам
func (o *OsmService) containsValidTags(tags map[string]string, group map[string][]string) bool {
	for _, list := range group {
		if o.matchTagsAgainstCompulsoryTagList(tags, list) {
			return true
		}
	}
	return false
}

// Очищает теги от лишних пробелов
func (o *OsmService) trimTags(tags map[string]string) map[string]string {
	trimmed := make(map[string]string)
	for k, v := range tags {
		trimmed[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return trimmed
}

// Проверяет наличие тегов у объекта
func (o *OsmService) hasTags(tags map[string]string) bool {
	n := len(tags)
	if n == 0 {
		return false
	}
	return true
}

// Проверяет наличие конкретных тегов у объекта по спискам
func (o *OsmService) matchTagsAgainstCompulsoryTagList(tags map[string]string, tagList []string) bool {
	for _, name := range tagList {

		feature := strings.Split(name, "~")
		foundVal, foundKey := tags[feature[0]]

		// Проверка наличия тега в списке
		if !foundKey {
			return false
		}

		// Проверка значения тега
		if len(feature) > 1 {
			if foundVal != feature[1] {
				return false
			}
		}
	}

	return true
}

// Получить значение тега по названию
func (o *OsmService) getTagByName(tags map[string]string, name string) string {
	foundVal, foundKey := tags[name]
	// Проверка наличия тега в списке
	if !foundKey {
		return ""
	}

	return foundVal
}

// Проверяет наличие ошибки и логирует ее
func (o *OsmService) checkFatalError(err error) {
	if err != nil {
		o.logger.Fatal(err.Error())
		os.Exit(1)
	}
}
