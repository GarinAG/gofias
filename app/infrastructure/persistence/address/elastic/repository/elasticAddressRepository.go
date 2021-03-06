package repository

import (
	"context"
	"encoding/json"
	cache "github.com/AeroAgency/golang-bigcache-lib"
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
          "district_kladr": {
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
          "area_guid": {
            "type": "keyword"
          },
          "area_kladr": {
            "type": "keyword"
          },
          "area": {
            "type": "keyword"
          },
          "area_type": {
            "type": "keyword"
          },
          "area_full": {
            "type": "keyword"
          },
          "city_guid": {
            "type": "keyword"
          },
          "city_kladr": {
            "type": "keyword"
          },
          "city": {
            "type": "keyword"
          },
          "city_type": {
            "type": "keyword"
          },
          "city_full": {
            "type": "keyword"
          },
          "settlement_guid": {
            "type": "keyword"
          },
          "settlement_kladr": {
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
          "street_guid": {
            "type": "keyword"
          },
          "street_kladr": {
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
	logger        interfaces.LoggerInterface // Логгер
	batchSize     int                        // Размер пачки для обновления
	elasticClient *elasticHelper.Client      // Клиент эластика
	indexName     string                     // Название индекса
	jobs          chan dto.JsonAddressDto    // Список задач для индексации
	results       chan dto.JsonAddressDto    // Список объектов индексации
	noOfWorkers   int                        // Количество обработчиков индексации
	indexCache    cache.CacheInterface       // Кэш объектов индексации
}

// Инициализация репозитория
func NewElasticAddressRepository(elasticClient *elasticHelper.Client, logger interfaces.LoggerInterface, batchSize int, prefix string, noOfWorkers int, cache cache.CacheInterface) repository.AddressRepositoryInterface {
	if noOfWorkers == 0 {
		noOfWorkers = 5
	}

	return &ElasticAddressRepository{
		logger:        logger,
		elasticClient: elasticClient,
		batchSize:     batchSize,
		indexName:     prefix + entity.AddressObject{}.TableName(),
		noOfWorkers:   noOfWorkers,
		indexCache:    cache,
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

// Найти адреса по GUID
func (a *ElasticAddressRepository) GetAddressByGuidList(guids []string) ([]*entity.AddressObject, error) {
	if len(guids) == 0 {
		return nil, nil
	}
	// Инициализирует сервис выборки элементов через ScrollApi
	scrollService := a.elasticClient.Client.Scroll(a.GetIndexName()).
		Query(elastic.NewTermsQuery("ao_guid", util.ConvertStringSliceToInterface(guids)...))

	scrollData, err := a.elasticClient.ScrollData(scrollService, a.batchSize)
	if err != nil {
		a.logger.Error(err.Error())
	}

	var items []*entity.AddressObject
	var item *dto.JsonAddressDto

	// Получает данные из эластика пачками
	for _, hit := range scrollData {
		// Конвертирует структуру ответа в DTO
		if err := json.Unmarshal(hit.Source, &item); err != nil {
			a.logger.Fatal(err.Error())
		}
		items = append(items, item.ToEntity())
	}

	return items, nil
}

// Найти адрес по GUID
func (a *ElasticAddressRepository) GetByGuid(guid string) (*entity.AddressObject, error) {
	if guid == "" {
		return nil, nil
	}
	res, err := a.GetAddressByGuidList([]string{guid})
	if err != nil || res == nil {
		return nil, err
	}

	return res[0], nil
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
	// Инициализирует сервис выборки элементов через ScrollApi
	scrollService := a.elasticClient.Client.Scroll(a.GetIndexName()).
		Query(elastic.NewBoolQuery().Filter(
			elastic.NewTermQuery("short_name", "г"),
			elastic.NewTermsQuery("ao_level", 1, 4))).
		Sort("ao_level", true)

	scrollData, err := a.elasticClient.ScrollData(scrollService, a.batchSize)
	if err != nil {
		a.logger.Error(err.Error())
	}

	var items []*entity.AddressObject
	var item *dto.JsonAddressDto

	// Получает данные из эластика пачками
	for _, hit := range scrollData {
		// Конвертирует структуру ответа в DTO
		if err := json.Unmarshal(hit.Source, &item); err != nil {
			a.logger.Fatal(err.Error())
		}
		items = append(items, item.ToEntity())
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
func (a *ElasticAddressRepository) GetAddressByTerm(term string, size int64, from int64, filter ...entity.FilterObject) ([]*entity.AddressObject, error) {
	if size == 0 {
		size = 100
	}
	queries := []elastic.Query{elastic.NewMatchQuery("address_suggest", term).Operator("and")}
	queries = a.prepareFilter(queries, filter...)

	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewBoolQuery().Must(queries...)).
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

// Подготовить фильтр для запроса
func (a *ElasticAddressRepository) prepareFilter(queries []elastic.Query, filters ...entity.FilterObject) []elastic.Query {
	for _, filter := range filters {
		if len(filter.Level.Values) > 0 {
			queries = append(queries, elastic.NewTermsQuery("ao_level", util.ConvertFloat32SliceToInterface(filter.Level.Values)...))
		}
		if filter.Level.Min > 0 || filter.Level.Max > 0 {
			rangeQuery := elastic.NewRangeQuery("ao_level")
			if filter.Level.Min > 0 {
				rangeQuery.Gte(filter.Level.Min)
			}
			if filter.Level.Max > 0 {
				rangeQuery.Lte(filter.Level.Max)
			}
			queries = append(queries, rangeQuery)
		}
		if len(filter.ParentGuid.Values) > 0 {
			queries = append(queries, elastic.NewTermsQuery("parent_guid", util.ConvertStringSliceToInterface(filter.ParentGuid.Values)...))
		}
		if len(filter.KladrId.Values) > 0 {
			queries = append(queries, elastic.NewTermsQuery("code", util.ConvertStringSliceToInterface(filter.KladrId.Values)...))
		}
	}

	return queries
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

// Найти адрес по почтовому индексу
func (a *ElasticAddressRepository) GetNearestCity(lon float64, lat float64) (*entity.AddressObject, error) {
	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewBoolQuery().Filter(
			elastic.NewGeoDistanceQuery("location").Lon(lon).Lat(lat).Distance("20km"),
			elastic.NewRangeQuery("ao_level").Lte(6))).
		Size(1).
		SortBy(
			elastic.NewGeoDistanceSort("location").
				Asc().
				Point(lat, lon).
				SortMode("min")).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	var item *dto.JsonAddressDto
	// Конвертирует структуру ответа в DTO
	if len(res.Hits.Hits) > 0 {
		for _, el := range res.Hits.Hits {
			if err := json.Unmarshal(el.Source, &item); err != nil {
				return nil, err
			}
			return item.ToEntity(), nil
		}
	}

	return nil, nil
}

// Найти адрес по почтовому индексу
func (a *ElasticAddressRepository) GetNearestAddress(lon float64, lat float64, term string) (*entity.AddressObject, error) {
	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewBoolQuery().Must(
			elastic.NewMatchQuery("address_suggest", term).Operator("and")).Filter(
			elastic.NewGeoDistanceQuery("location").Lon(lon).Lat(lat).Distance("5km"))).
		Size(1).
		SortBy(
			elastic.NewGeoDistanceSort("location").
				Asc().
				Point(lat, lon).
				SortMode("min")).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	var item *dto.JsonAddressDto
	// Конвертирует структуру ответа в DTO
	if len(res.Hits.Hits) > 0 {
		for _, el := range res.Hits.Hits {
			if err := json.Unmarshal(el.Source, &item); err != nil {
				return nil, err
			}
			return item.ToEntity(), nil
		}
	}

	return nil, nil
}

// Получить объект для работы с пачками элементов
func (a *ElasticAddressRepository) GetBulkService() *elastic.BulkService {
	return a.elasticClient.Client.Bulk().Index(a.GetIndexName())
}

// Обновить коллекцию адресов
func (a *ElasticAddressRepository) InsertUpdateCollection(wg *sync.WaitGroup, channel <-chan interface{}, count chan<- int, isFull bool) {
	defer wg.Done()
	begin := time.Now()
	var total uint64
	step := 1
	var deleted []string
	updated := make(map[string]dto.JsonAddressDto)

	// Цикл получения объекта адреса из канала
	for d := range channel {
		if d == nil {
			break
		}
		total++
		saveItem := dto.JsonAddressDto{}
		saveItem.GetFromEntity(d.(entity.AddressObject))
		// Проверяет активность объекта
		if saveItem.IsActive() {
			// Добавляет объект в очередь на сохранение
			updated[saveItem.AoGuid] = saveItem
		} else {
			// Добавляет объект в очередь на удаление
			deleted = append(deleted, saveItem.ID)
		}

		// Отправляет запросы в эластик при превышении размера пачки
		if len(updated)+len(deleted) >= a.batchSize {
			a.update(updated, deleted, isFull)
			deleted = nil
			if total%uint64(100000) == 0 && !util.CanPrintProcess {
				a.logger.WithFields(interfaces.LoggerFields{"step": step, "count": total}).Info("Add addresses to index")
				step++
			}
		}
	}

	// Отправляет оставшиеся запросы в эластик
	if len(updated)+len(deleted) > 0 {
		a.update(updated, deleted, isFull)
		deleted = nil
	}
	if !util.CanPrintProcess {
		a.logger.WithFields(interfaces.LoggerFields{"step": step, "count": total}).Info("Add addresses to index")
	}
	a.logger.WithFields(interfaces.LoggerFields{"count": total, "execTime": humanize.RelTime(begin, time.Now(), "", "")}).Info("Address import execution time")
	a.Refresh()
	count <- int(total)
}

// Сохраняет данные в эластик
func (a *ElasticAddressRepository) update(updated map[string]dto.JsonAddressDto, deleted []string, isFull bool) {
	bulk := a.GetBulkService()
	ctx := context.Background()
	// Дополняет элементы полями из БД
	if len(updated) > 0 && !isFull {
		var updatedKeys []string
		for k := range updated {
			updatedKeys = append(updatedKeys, k)
		}

		items, _ := a.GetAddressByGuidList(updatedKeys)
		for _, item := range items {
			updateItem, ok := updated[item.AoGuid]
			if ok {
				updateItem.UpdateFromExistItem(*item)
				updated[item.AoGuid] = updateItem
			}
		}
	}
	for k, item := range updated {
		bulk.Add(elastic.NewBulkIndexRequest().Id(item.ID).Doc(item))
		delete(updated, k)
	}
	for _, item := range deleted {
		bulk.Add(elastic.NewBulkDeleteRequest().Id(item))
	}

	res, err := bulk.Do(ctx)
	if err != nil {
		a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Add addresses bulk commit failed")
	}
	if res != nil && res.Errors {
		a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("Add addresses bulk commit failed")
	}
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
		if len(guids) > 0 {
			// Добавляет фильтр на ограничение выборки по списку GUID
			guidsInterface := util.ConvertStringSliceToInterface(guids)
			queries = append(queries, elastic.NewTermsQuery("ao_guid", guidsInterface...))
		} else {
			// Добавляет фильтр на ограничение выборки по дате начала импорта
			queries = append(queries, elastic.NewRangeQuery("bazis_update_date").Gte(start.Format(util.TimeFormat)))
		}
	} else {
		// Индексирует все элементы в индексе
		a.logger.Info("Full indexing...")
	}
	query = elastic.NewBoolQuery().Must(queries...)

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
	scrollService.Scroll("10m")
	count := 0

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
		// Добавляет элементы в пул задач
		for _, hit := range res.Hits.Hits {
			var item dto.JsonAddressDto
			// Конвертирует структуру ответа в DTO
			if err := json.Unmarshal(hit.Source, &item); err != nil {
				a.logger.Fatal(err.Error())
			}
			a.jobs <- item
		}
	}

	// Принудительно закрывает сервис выборки элементов
	err := scrollService.Clear(ctx)
	if err != nil {
		a.logger.Error(err.Error())
	}
	a.logger.WithFields(interfaces.LoggerFields{"count": count}).Info("Address update count")

	close(a.jobs)
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
		guid := address.ParentGuid
		// Формирует информацию об адресе объекта
		address.FullName = util.PrepareFullName(address.ShortName, address.FormalName)
		address.FullAddress = address.FullName
		address.AddressSuggest = util.PrepareSuggest("", address.ShortName, address.FormalName)

		// Ищет родительские объекты и дополняет адрес текущего объекта
		if guid != "" {
			var search *entity.AddressObject
			searchObject := entity.AddressObject{}
			// Ищет родительский объект в кэше
			searchResult := a.indexCache.Get(guid, &searchObject)
			if searchResult == nil {
				search, _ = a.GetByGuid(guid)
			} else {
				search = searchResult.(*entity.AddressObject)
			}

			if search != nil {
				// Конвертирует объект адреса в DTO
				dtoItem.GetFromEntity(*search)
				// Дополняет адрес текущего объекта
				if address.AoLevel == 4 {
					// Пропускаем район в названии
					address.FullAddress = dtoItem.RegionFull + ", " + address.FullAddress
				} else {
					address.FullAddress = dtoItem.FullAddress + ", " + address.FullAddress
				}
				address.AddressSuggest = dtoItem.AddressSuggest + ", " + address.AddressSuggest
				// Формирует информацию о регионе объекта
				if dtoItem.Region != "" {
					address.RegionGuid = dtoItem.RegionGuid
					address.RegionKladr = dtoItem.RegionKladr
					address.Region = dtoItem.Region
					address.RegionType = dtoItem.RegionType
					address.RegionFull = dtoItem.RegionFull
				}
				// Формирует информацию о районе объекта
				if dtoItem.Area != "" {
					address.AreaGuid = dtoItem.AreaGuid
					address.AreaKladr = dtoItem.AreaKladr
					address.Area = dtoItem.Area
					address.AreaType = dtoItem.AreaType
					address.AreaFull = dtoItem.AreaFull
				}
				// Формирует информацию о городе объекта
				if dtoItem.City != "" {
					address.CityGuid = dtoItem.CityGuid
					address.CityKladr = dtoItem.CityKladr
					address.City = dtoItem.City
					address.CityType = dtoItem.CityType
					address.CityFull = dtoItem.CityFull
				}
				// Устанавливает населенный пункт объекта
				if dtoItem.Settlement != "" {
					address.SettlementGuid = dtoItem.SettlementGuid
					address.SettlementKladr = dtoItem.SettlementKladr
					address.Settlement = dtoItem.Settlement
					address.SettlementType = dtoItem.SettlementType
					address.SettlementFull = dtoItem.SettlementFull
				}
			}
		}

		if address.AoLevel <= 2 {
			address.RegionGuid = address.AoGuid
			address.RegionKladr = address.Code
			address.Region = strings.TrimSpace(address.FormalName)
			address.RegionType = strings.TrimSpace(address.ShortName)
			address.RegionFull = util.PrepareFullName(address.RegionType, address.Region)
		} else if address.AoLevel == 3 {
			address.AreaGuid = address.AoGuid
			address.AreaKladr = address.Code
			address.Area = strings.TrimSpace(address.FormalName)
			address.AreaType = strings.TrimSpace(address.ShortName)
			if address.RegionFull != "" {
				address.AreaFull = address.RegionFull + ", "
			}
			address.AreaFull += util.PrepareFullName(address.AreaType, address.Area)
		} else if address.AoLevel == 4 {
			address.CityGuid = address.AoGuid
			address.CityKladr = address.Code
			address.City = strings.TrimSpace(address.FormalName)
			address.CityType = strings.TrimSpace(address.ShortName)
			if address.AreaFull != "" {
				address.CityFull = address.AreaFull + ", "
			} else if address.RegionFull != "" {
				address.CityFull = address.RegionFull + ", "
			}
			address.CityFull += util.PrepareFullName(address.CityType, address.City)
		} else if address.AoLevel == 5 || address.AoLevel == 6 {
			address.SettlementGuid = address.AoGuid
			address.SettlementKladr = address.Code
			address.Settlement = strings.TrimSpace(address.FormalName)
			address.SettlementType = strings.TrimSpace(address.ShortName)
			address.SettlementFull = ""
			address.SettlementFull = ""
			if address.CityFull != "" {
				address.SettlementFull = address.CityFull + ", "
			} else if address.AreaFull != "" {
				address.SettlementFull = address.AreaFull + ", "
			} else if address.RegionFull != "" {
				address.SettlementFull = address.RegionFull + ", "
			}
			address.SettlementFull += util.PrepareFullName(address.SettlementType, address.Settlement)
		} else if address.AoLevel == 7 {
			address.StreetGuid = address.AoGuid
			address.StreetKladr = address.Code
			address.StreetType = strings.TrimSpace(address.ShortName)
			address.Street = strings.TrimSpace(address.FormalName)
			address.StreetFull = ""
			if address.SettlementFull != "" {
				address.StreetFull = address.SettlementFull + ", "
			} else if address.CityFull != "" {
				address.StreetFull = address.CityFull + ", "
			} else if address.AreaFull != "" {
				address.StreetFull = address.AreaFull + ", "
			} else if address.RegionFull != "" {
				address.StreetFull = address.RegionFull + ", "
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
	bar := util.StartNewProgress(int(total), "Indexing addresses", false)

	for d := range a.results {
		// Добавляет объект в индексацию домов, если данный объект является улицей
		if d.AoLevel == 7 && indexChan != nil {
			indexChan <- entity.IndexObject{
				AoGuid:         d.AoGuid,
				FullAddress:    d.FullAddress,
				AddressSuggest: d.AddressSuggest,
			}
		} else if d.AoLevel <= 6 {
			a.indexCache.Set(d.AoGuid, d.ToEntity())
		}

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
			a.clearIndexCache()
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
		a.clearIndexCache()
	}
	bar.Finish()
	a.logger.WithFields(interfaces.LoggerFields{"execTime": humanize.RelTime(begin, time.Now(), "", "")}).Info("Address index execution time")
	done <- true
	if indexChan != nil {
		close(indexChan)
	}
}

// Очистка кеша
func (a *ElasticAddressRepository) clearIndexCache() {
	// Обновляет индекс
	a.Refresh()
	// Очищает кэш
	a.indexCache.Clear()
}
