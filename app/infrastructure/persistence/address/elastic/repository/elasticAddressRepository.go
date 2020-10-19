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
	// Структура индекса в эластике
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
                "filter": ["lowercase", "edge_ngram"],
                "tokenizer": "standard"
              },
              "keyword_analyzer": {
                "filter": ["lowercase"],
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
            "type": "keyword"
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
          "district_guid": {
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
          "settlement_guid": {
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
            "type": "geo_point",
            "ignore_malformed": true
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
	// Обработчик удаления старых адресов
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

// Репозиторий адресов в эластике
type ElasticAddressRepository struct {
	logger        interfaces.LoggerInterface       // Логгер
	batchSize     int                              // Размер пачки для обновления
	elasticClient *elasticHelper.Client            // Клиент эластика
	indexName     string                           // Название индекса
	jobs          chan dto.JsonAddressDto          // Список задач для индексации
	results       chan dto.JsonAddressDto          // Список объектов индексации
	noOfWorkers   int                              // Количество обработчиков индексации
	indexCache    map[string]*entity.AddressObject // Кэш объектов индексации
	indexMutex    sync.RWMutex                     // Объект блокировки записи в кэш
}

// Инициализация репозитория
func NewElasticAddressRepository(elasticClient *elasticHelper.Client, logger interfaces.LoggerInterface, batchSize int, prefix string, noOfWorkers int) repository.AddressRepositoryInterface {
	if noOfWorkers == 0 {
		noOfWorkers = 5
	}

	return &ElasticAddressRepository{
		logger:        logger,
		elasticClient: elasticClient,
		batchSize:     batchSize,
		indexName:     prefix + entity.AddressObject{}.TableName(),
		noOfWorkers:   noOfWorkers,
	}
}

// Инициализация индекса
func (a *ElasticAddressRepository) Init() error {
	// Создание индекса
	err := a.elasticClient.CreateIndex(a.indexName, addrIndexSettings)
	if err != nil {
		return err
	}

	// Добавление процессора для удаления старых объектов
	return a.elasticClient.CreatePreprocessor(addrPipelineId, addrDropPipeline)
}

// Получить назваине индекса
func (a *ElasticAddressRepository) GetIndexName() string {
	return a.indexName
}

// Удалить индекс
func (a *ElasticAddressRepository) Clear() error {
	return a.elasticClient.DropIndex(a.indexName)
}

// Найти адрес по названию
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
	// Конвертирует структуру ответа в DTO
	if len(res.Hits.Hits) > 0 {
		if err := json.Unmarshal(res.Hits.Hits[0].Source, &item); err != nil {
			return nil, err
		}

		return item.ToEntity(), nil
	}

	return nil, nil
}

// Найти адрес по GUID
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
	// Конвертирует структуру ответа в DTO
	if len(res.Hits.Hits) > 0 {
		if err := json.Unmarshal(res.Hits.Hits[0].Source, &item); err != nil {
			return nil, err
		}

		return item.ToEntity(), nil
	}

	return nil, nil
}

// Найти город по названию
func (a *ElasticAddressRepository) GetCityByFormalName(term string) (*entity.AddressObject, error) {
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
	// Конвертирует структуру ответа в DTO
	if len(res.Hits.Hits) > 0 {
		if err := json.Unmarshal(res.Hits.Hits[0].Source, &item); err != nil {
			return nil, err
		}

		return item.ToEntity(), nil
	}

	return nil, nil
}

// Подсчитать количество адресов по фильтру
func (a *ElasticAddressRepository) CountAllData(query interface{}) (int64, error) {
	if query == nil {
		query = elastic.NewBoolQuery()
	}
	return a.elasticClient.CountAllData(a.GetIndexName(), query.(elastic.Query))
}

// Получить список всех городов
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
	// Конвертирует структуру ответа в DTO
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

// Найти города по подстроке
func (a *ElasticAddressRepository) GetCitiesByTerm(term string, size int64, from int64) ([]*entity.AddressObject, error) {
	if size == 0 {
		size = 100
	}

	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewBoolQuery().Must(
			elastic.NewMultiMatchQuery(term, "address_suggest").Operator("and")).
			Filter(elastic.NewTermsQuery("ao_level", 1, 4))).
		From(int(from)).
		Size(int(size)).
		Sort("ao_level", true).
		Sort("full_address", true).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	var items []*entity.AddressObject
	var item *dto.JsonAddressDto
	// Конвертирует структуру ответа в DTO
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

// Найти адрес по подстроке
func (a *ElasticAddressRepository) GetAddressByTerm(term string, size int64, from int64) ([]*entity.AddressObject, error) {
	if size == 0 {
		size = 100
	}

	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewBoolQuery().Must(
			elastic.NewMatchQuery("address_suggest", term).Operator("and"))).
		From(int(from)).
		Size(int(size)).
		Sort("ao_level", true).
		Sort("_score", false).
		Sort("full_address", true).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	var items []*entity.AddressObject
	var item *dto.JsonAddressDto
	// Конвертирует структуру ответа в DTO
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

// Найти адрес по почтовому индексу
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
		Sort("full_address", true).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	var items []*entity.AddressObject
	var item *dto.JsonAddressDto
	// Конвертирует структуру ответа в DTO
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

// Получить объект для работы с пачками элементов
func (a *ElasticAddressRepository) GetBulkService() *elastic.BulkService {
	return a.elasticClient.Client.Bulk().Index(a.GetIndexName())
}

// Обновить коллекцию адресов
func (a *ElasticAddressRepository) InsertUpdateCollection(channel <-chan interface{}, done <-chan bool, count chan<- int, isFull bool) {
	bulk := a.GetBulkService()
	ctx := context.Background()
	begin := time.Now()
	var total uint64
	step := 1

	// Цикл получения объекта адреса из канала
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
			// Проверяет активность объекта
			if saveItem.IsActive() {
				// При неполном импорте дополняет данными из индекса
				if !isFull {
					dbItem, _ := a.GetByGuid(saveItem.AoGuid)
					if dbItem != nil {
						saveItem.UpdateFromExistItem(*dbItem)
					}
				}
				// Добавляет объект в очередь на сохранение
				bulk.Add(elastic.NewBulkIndexRequest().Id(saveItem.ID).Doc(saveItem))
			} else {
				// Добавляет объект в очередь на удаление
				bulk.Add(elastic.NewBulkDeleteRequest().Id(saveItem.ID))
			}

			// Отправляет запросы в эластик при превышении размера пачки
			if bulk.NumberOfActions() >= a.batchSize {
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

	// Отправляет оставшиеся запросы в эластик
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

// Обновить индекс
func (a *ElasticAddressRepository) Refresh() {
	a.elasticClient.RefreshIndexes([]string{a.GetIndexName()})
}

// Переоткрыть индекс
func (a *ElasticAddressRepository) ReopenIndex() {
	a.elasticClient.Client.CloseIndex(a.GetIndexName())
	a.elasticClient.Client.OpenIndex(a.GetIndexName())
}

// Индексация адресов
func (a *ElasticAddressRepository) Index(isFull bool, start time.Time, guids []string, indexChan chan<- entity.IndexObject) error {
	done := make(chan bool)
	a.indexMutex = sync.RWMutex{}
	// Создает экземпляр кэша
	a.indexCache = make(map[string]*entity.AddressObject)
	// Создает канал для работы с объектами
	a.jobs = make(chan dto.JsonAddressDto, a.noOfWorkers)
	// Создает канал для сохранения объектов в индекс
	a.results = make(chan dto.JsonAddressDto, a.noOfWorkers)
	// Обновляет индекс
	a.Refresh()
	// Подготавливает фильтр для получения элементов
	query := a.prepareIndexQuery(isFull, start, guids)
	// Получает общее количество элементов по фильтру
	queryCount := a.calculateIndexCount(query)
	// Получает элементы из индекса для переиндексации
	go a.getIndexItems(query)
	// Обновляет элементы в индексе
	go a.saveIndexItems(done, time.Now(), queryCount, indexChan)
	// Создает пул задач на обработку элементов
	a.createWorkerPool(a.noOfWorkers)
	<-done
	// Обновляет индекс
	a.Refresh()

	return nil
}

// Подготовить фильтр для получения элементов
func (a *ElasticAddressRepository) prepareIndexQuery(isFull bool, start time.Time, guids []string) elastic.Query {
	var query elastic.Query
	var queries []elastic.Query
	// Проверяет, является ли индексация полной
	if !isFull {
		a.logger.Info("Indexing...")
		// Добавляет фильтр на ограничение выборки по уровню адреса
		queries = append(queries, elastic.NewRangeQuery("ao_level").Gt(1))
		// Добавляет фильтр на ограничение выборки по дате начала импорта
		queries = append(queries, elastic.NewRangeQuery("bazis_update_date").Gte(start.Format("2006-01-02")+"T00:00:00Z"))
		if len(guids) > 0 {
			// Добавляет фильтр на ограничение выборки по списку GUID
			guidsInterface := util.ConvertStringSliceToInterface(guids)
			query = elastic.NewBoolQuery().Should(elastic.NewBoolQuery().Must(queries...), elastic.NewBoolQuery().Must(elastic.NewTermsQuery("ao_guid", guidsInterface...)))
		} else {
			query = elastic.NewBoolQuery().Must(queries...)
		}
	} else {
		a.logger.Info("Full indexing...")
		// Индексирует все элементы в индексе
		query = elastic.NewBoolQuery().Must(queries...)
	}

	return query
}

// Получить общее количество элементов по фильтру
func (a *ElasticAddressRepository) calculateIndexCount(query elastic.Query) int64 {
	// Получает общее количество элементов
	addTotalCount, err := a.CountAllData(nil)
	if err != nil {
		a.logger.Error(err.Error())
	}
	// Получает количество элементов по фильтру
	queryCount, err := a.CountAllData(query)
	if err != nil {
		a.logger.Error(err.Error())
	}

	a.logger.WithFields(interfaces.LoggerFields{"count": addTotalCount}).Info("Total address count")
	a.logger.WithFields(interfaces.LoggerFields{"count": queryCount}).Info("Number of indexed addresses")

	return queryCount
}

// Получить элементы из индекса для переиндексации
func (a *ElasticAddressRepository) getIndexItems(query elastic.Query) {
	batch := a.batchSize
	// Ограничивает размер пачки при поиске
	if batch > 10000 {
		batch = 10000
	}

	// Инициализирует сервис выборки элементов через ScrollApi
	scrollService := a.elasticClient.Client.Scroll(a.GetIndexName()).
		Query(query).
		Sort("ao_level", true).
		Size(batch)

	ctx := context.Background()
	scrollService.Scroll("1s")
	count := 0
	var wg sync.WaitGroup

	// Получает данные из эластика пачками
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
		// Добавляет элементы в пул задач
		go a.addJobs(res.Hits.Hits, &wg)
	}

	// Принудительно закрывает сервис выборки элементов
	err := scrollService.Clear(ctx)
	if err != nil {
		a.logger.Error(err.Error())
	}

	wg.Wait()
	a.logger.WithFields(interfaces.LoggerFields{"count": count}).Info("Address update count")

	close(a.jobs)
}

// Добавить элементы в пул задач
func (a *ElasticAddressRepository) addJobs(hits []*elastic.SearchHit, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, hit := range hits {
		var item dto.JsonAddressDto
		// Конвертирует структуру ответа в DTO
		if err := json.Unmarshal(hit.Source, &item); err != nil {
			a.logger.Fatal(err.Error())
		}
		a.jobs <- item
	}
}

// Создать пул задач на обработку элементов
func (a *ElasticAddressRepository) createWorkerPool(noOfWorkers int) {
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		// Подготавливает элементы перед сохранением в индекс
		go a.prepareItemsBeforeSave(&wg)
	}
	wg.Wait()
	close(a.results)
}

// Подготовить элементы перед сохранением в индекс
func (a *ElasticAddressRepository) prepareItemsBeforeSave(wg *sync.WaitGroup) {
	for address := range a.jobs {
		// Устанавливает время обновления объекта
		address.UpdateBazisDate()
		dtoItem := dto.JsonAddressDto{}
		city := dto.JsonAddressDto{}
		district := dto.JsonAddressDto{}
		guid := address.ParentGuid
		// Формирует информацию об адресе объекта
		address.FullName = util.PrepareFullName(address.ShortName, address.FormalName)
		address.FullAddress = address.FullName
		address.AddressSuggest = util.PrepareSuggest("", address.ShortName, address.FormalName)

		// Ищет родительские объекты и дополняет адрес текущего объекта
		for guid != "" {
			// Блокирует объект кэша для поиска элемента
			a.indexMutex.RLock()
			// Ищет родительский объект в кэше
			search, ok := a.indexCache[guid]
			// Снимает блокировку кэша
			a.indexMutex.RUnlock()
			if !ok {
				search, _ = a.GetByGuid(guid)
			}

			if search != nil {
				// Конвертирует объект адреса в DTO
				dtoItem.GetFromEntity(*search)
				guid = dtoItem.ParentGuid
				// Дополняет адрес текущего объекта
				address.FullAddress = util.PrepareFullName(dtoItem.ShortName, dtoItem.FormalName) + ", " + address.FullAddress
				address.AddressSuggest = util.PrepareSuggest(address.AddressSuggest, dtoItem.ShortName, dtoItem.FormalName)

				// Устанавливает город объекта
				if dtoItem.AoLevel >= 4 {
					city = dtoItem
				}
				// Устанавливает район объекта
				if dtoItem.AoLevel < 4 {
					district = dtoItem
				}
			} else {
				guid = ""
			}
		}

		// Формирует информацию о районе объекта
		address.DistrictGuid = district.AoGuid
		address.District = strings.TrimSpace(district.FormalName)
		address.DistrictType = strings.TrimSpace(district.ShortName)
		if address.District != "" {
			address.DistrictFull = util.PrepareFullName(address.DistrictType, address.District)
		}

		// Формирует информацию о поселении объекта
		address.SettlementGuid = city.AoGuid
		address.Settlement = strings.TrimSpace(city.FormalName)
		address.SettlementType = strings.TrimSpace(city.ShortName)
		if address.Settlement != "" {
			address.SettlementFull = ""
			if address.DistrictFull != "" {
				address.SettlementFull = address.DistrictFull + ", "
			}
			address.SettlementFull += util.PrepareFullName(address.SettlementType, address.Settlement)
		}

		// Формирует информацию об улице объекта
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

		a.results <- address
	}

	wg.Done()
}

// Обновить элементы в индексе
func (a *ElasticAddressRepository) saveIndexItems(done chan bool, begin time.Time, total int64, indexChan chan<- entity.IndexObject) {
	// Получает объект для работы с пачками элементов
	bulk := a.GetBulkService()
	ctx := context.Background()
	// Инициализация прогресс-бара
	bar := util.StartNewProgress(int(total))

	for d := range a.results {
		// Добавляет объект в индексацию домов, если данный объект является улицей
		if d.AoLevel == 7 && indexChan != nil {
			indexChan <- entity.IndexObject{
				AoGuid:         d.AoGuid,
				FullAddress:    d.FullAddress,
				AddressSuggest: d.AddressSuggest,
			}
		}

		// Блокирует кэш для записи
		a.indexMutex.Lock()
		// Записывает объект в кэш
		a.indexCache[d.AoGuid] = d.ToEntity()
		// Снимает блокировку кэша
		a.indexMutex.Unlock()

		// Добавляет объект в очередь на сохранение
		bulk.Add(elastic.NewBulkIndexRequest().Id(d.ID).Doc(d))
		bar.Increment()
		// Отправляет запросы в эластик при превышении размера пачки
		if bulk.NumberOfActions() >= a.batchSize {
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

	// Отправляет оставшиеся запросы в эластик
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
	close(indexChan)
}
