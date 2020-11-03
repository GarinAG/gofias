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
	"sync"
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
	file, err := o.downloadService.DownloadFile(o.config.GetConfig().Osm.Url, "russia.pbf")
	o.checkFatalError(err)
	if file != nil {
		// Разбирает файл с данными
		o.parseFile(file.Path)
	}
}

// Разбирает файл с данными
func (o *OsmService) parseFile(filepath string) {
	o.logger.Info("Start parsing OSM file")

	// Открывает файл с данными OSM
	f, err := os.Open(filepath)
	o.checkFatalError(err)
	defer f.Close()

	// Создает объект сканнера
	scanner := osmpbf.New(context.Background(), f, 3)
	defer scanner.Close()

	addressChan := make(chan *entity.Node)
	housesChan := make(chan *entity.Node)
	// Проверяет наличие домов в БД
	housesCnt, _ := o.houseRepo.CountAllData(nil)
	housesCnt = 0 // TODO Enable houses import
	if housesCnt == 0 {
		housesChan = nil
	}

	var wg sync.WaitGroup
	wg.Add(2)
	// Сканирует файл с данными OSM
	go o.scan(&wg, scanner, addressChan, housesChan)
	// Обновляет адреса
	go o.updateAddresses(&wg, addressChan)
	// При наличии домов разрешает обновление местоположений
	if housesChan != nil {
		wg.Add(1)
		// Обновляет дома
		go o.updateHouses(&wg, housesChan)
	}
	wg.Wait()
	o.downloadService.ClearDirectory()
	o.logger.Info("OSM parsing finished")
	scanErr := scanner.Err()
	if scanErr != nil {
		panic(scanErr)
	}
}

// Сканирует файл с данными OSM
func (o *OsmService) scan(wg *sync.WaitGroup, scanner *osmpbf.Scanner, addressChan chan<- *entity.Node, housesChan chan<- *entity.Node) {
	defer wg.Done()
	// Создает список условий, проставляет обязательное условие - наличие названия
	tagList := "name"
	conditions := make(map[string][]string)
	for _, group := range strings.Split(tagList, ",") {
		conditions[group] = strings.Split(group, "+")
	}

	bar := util.StartNewProgress(-1, "Import OSM", false)

	for scanner.Scan() {
		switch e := scanner.Object().(type) {
		// Элемент является объектом
		case *osm.Node:
			if e.Tags != nil {
				// Проверяет условия
				if o.hasTags(e.TagMap()) && o.containsValidTags(e.TagMap(), conditions) {
					// Проверяет и разбирает объект
					node := o.prepareItems(e)
					bar.Increment()
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
			//case *osm.Way:
			//case *osm.Relation:
		}
	}

	bar.Finish()
	close(addressChan)
	if housesChan != nil {
		close(housesChan)
	}
}

// Обновляет адреса
func (o *OsmService) updateAddresses(wg *sync.WaitGroup, addressChan <-chan *entity.Node) {
	defer wg.Done()
	address := make(chan interface{})
	addressCnt := make(chan int)
	var importWg sync.WaitGroup
	importWg.Add(1)
	// Сохраняет элементы в БД
	go o.addressRepo.InsertUpdateCollection(&importWg, address, addressCnt, true)
	f, err := os.Create("lines.txt")
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}
	defer f.Close()

	for d := range addressChan {
		// Ищет адреса в БД по названию
		items, _ := o.addressRepo.GetAddressByTerm(d.Name, 1, 0)
		if len(items) > 0 {
			item := items[0]
			location := fmt.Sprint(d.Lat, ",", d.Lon)
			// Сохраняет только адреса, у которых отличается местоположение или индекс с данными из OSM
			if item.Location != location || (d.PostalCode != "" && item.PostalCode != d.PostalCode) {
				item.Location = location
				if d.PostalCode != "" && item.PostalCode != d.PostalCode {
					item.PostalCode = d.PostalCode
				}
				address <- *item
			}
		} else {
			fmt.Fprintln(f, d.Name, d.Lat, d.Lon)
		}
	}
	close(address)
	<-addressCnt
	importWg.Wait()
}

// Обновляет дома
func (o *OsmService) updateHouses(wg *sync.WaitGroup, housesChan <-chan *entity.Node) {
	defer wg.Done()
	houses := make(chan interface{})
	housesCnt := make(chan int)
	var importWg sync.WaitGroup
	importWg.Add(1)
	// Сохраняет элементы в БД
	go o.houseRepo.InsertUpdateCollection(&importWg, houses, housesCnt, true)

	for d := range housesChan {
		// Ищет ближайщий адрес при отсутствии города
		if d.HouseAddress != "" {
			nearest, _ := o.addressRepo.GetNearestCity(d.Lon, d.Lat)
			if nearest == nil {
				continue
			} else {
				d.Name = nearest.FullAddress + " " + d.Name
			}
		}

		// Ищет дома в БД по адресу
		items, _ := o.houseRepo.GetAddressByTerm(d.Name, 1, 0)
		if len(items) > 0 {
			item := items[0]
			location := fmt.Sprint(d.Lat, ",", d.Lon)
			// Сохраняет только дома, у которых отличается местоположение или индекс с данными из OSM
			if item.Location != location || (d.PostalCode != "" && item.PostalCode != d.PostalCode) {
				item.Location = location
				if d.PostalCode != "" && item.PostalCode != d.PostalCode {
					item.PostalCode = d.PostalCode
				}
				houses <- *item
			}
		}
	}
	close(houses)
	<-housesCnt
	importWg.Wait()
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
	houseAddress := ""
	if strings.Contains(district, "городской округ") {
		district = ""
	}

	// Проверяет наличие тегов у объекта
	if place != "" || (housenum != "" && street != "") {
		if place != "" && place != "city" && place != "town" && region == "" && district == "" && city == "" {
			return nil
		}

		fullAddr := ""
		if region != "" {
			fullAddr = region
		}
		if district != "" {
			if fullAddr != "" {
				fullAddr += ", "
			}
			fullAddr += district
		}
		if city != "" {
			if fullAddr != "" {
				fullAddr += ", "
			}
			if place == "city" || place == "town" {
				fullAddr += "город "
			}
			fullAddr += city
		}
		if street != "" {
			if fullAddr != "" {
				fullAddr += ", "
			}
			fullAddr += street
		}
		if housenum != "" {
			houseAddress = fullAddr
			if fullAddr != "" {
				fullAddr += ", "
			}
			fullAddr += housenum
		} else if name != "" {
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
			Node:       e,
		}

		if place != "" {
			node.Type = "place"
		} else {
			node.Type = "building"
			if city == "" {
				node.HouseAddress = houseAddress
			}
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
