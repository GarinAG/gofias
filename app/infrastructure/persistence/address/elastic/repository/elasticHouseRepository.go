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
	"sync"
	"time"
)

const (
	// Структура индекса в эластике
	houseIndexSettings = `
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
			  "ngram": {
				"type": "ngram",
				"min_gram": "1",
				"max_gram": "15"
			  },
			  "edge_ngram": {
                "type": "edge_ngram",
                "min_gram": "1",
                "max_gram": "50"
			  }
			},
			"analyzer": {
			  "ngram_analyzer": {
				"filter": ["lowercase", "ngram"],
				"tokenizer": "standard"
			  },
			  "edge_ngram_analyzer": {
				"filter": ["lowercase", "edge_ngram"],
				"tokenizer": "standard"
			  },
			  "keyword_analyzer": {
				"filter": ["lowercase"],
				"tokenizer": "standard"
			  }
			}
		  },
          "max_ngram_diff": 14
		}
	  },
	  "mappings": {
		"dynamic": false,
		"properties": {
		  "house_id": {
			"type": "keyword"
		  },
		  "house_guid": {
			"type": "keyword"
		  },
		  "ao_guid": {
			"type": "keyword"
		  },
		  "build_num": {
			"type": "keyword"
		  },
		  "house_num": {
			"type": "keyword"
		  },
          "address_suggest": {
            "type": "text",
            "analyzer": "edge_ngram_analyzer",
            "search_analyzer": "keyword_analyzer"
          },
		  "house_full_num": {
			"type": "text",
			"analyzer": "ngram_analyzer",
			"search_analyzer": "keyword_analyzer",
			"fields": {
			  "keyword": {
				"type": "keyword"
			  }
			}
		  },
		  "full_address": {
			"type": "keyword"
		  },
		  "str_num": {
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
		  "bazis_update_date": {
			"type": "date"
		  },
		  "update_date": {
			"type": "date"
		  },
		  "cad_num": {
			"type": "keyword"
		  },
		  "okato": {
			"type": "keyword"
		  },
		  "oktmo": {
			"type": "keyword"
		  },
		  "location": {
			"type": "geo_point",
            "ignore_malformed": true
		  }
		}
	  }
	}
	`
	// Обработчик удаления старых домов
	housesPipelineId  = "house_drop_pipeline"
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

// Репозиторий домов в эластике
type ElasticHouseRepository struct {
	elasticClient *elasticHelper.Client      // Клиент эластика
	logger        interfaces.LoggerInterface // Логгер
	batchSize     int                        // Размер пачки для обновления
	indexName     string                     // Название индекса
	results       chan dto.JsonHouseDto      // Список объектов индексации
	noOfWorkers   int                        // Количество обработчиков индексации
}

// Инициализация репозитория
func NewElasticHouseRepository(elasticClient *elasticHelper.Client, logger interfaces.LoggerInterface, batchSize int, prefix string, noOfWorkers int) repository.HouseRepositoryInterface {
	if noOfWorkers == 0 {
		noOfWorkers = 10
	}

	return &ElasticHouseRepository{
		elasticClient: elasticClient,
		logger:        logger,
		batchSize:     batchSize,
		indexName:     prefix + entity.HouseObject{}.TableName(),
		noOfWorkers:   noOfWorkers,
	}
}

// Инициализация индекса
func (a *ElasticHouseRepository) Init() error {
	// Создание индекса
	err := a.elasticClient.CreateIndex(a.indexName, houseIndexSettings)
	if err != nil {
		return err
	}

	// Добавление процессора для удаления старых объектов
	return a.elasticClient.CreatePreprocessor(housesPipelineId, houseDropPipeline)
}

// Получить назваине индекса
func (a *ElasticHouseRepository) GetIndexName() string {
	return a.indexName
}

// Удалить индекс
func (a *ElasticHouseRepository) Clear() error {
	return a.elasticClient.DropIndex(a.indexName)
}

// Получить элементы из индекса через ScrollApi
func (a *ElasticHouseRepository) scroll(scrollService *elastic.ScrollService) ([]*entity.HouseObject, error) {
	scrollData, err := a.elasticClient.ScrollData(scrollService, a.batchSize)
	if err != nil {
		a.logger.Error(err.Error())
	}

	var items []*entity.HouseObject
	var item *dto.JsonHouseDto

	// Получает данные из эластика пачками
	for _, hit := range scrollData {
		// Конвертирует структуру ответа в DTO
		if err := json.Unmarshal(hit.Source, &item); err != nil {
			return nil, err
		}
		items = append(items, item.ToEntity())
	}

	return items, nil
}

// Найти дом по GUID
func (a *ElasticHouseRepository) GetByGuid(guid string) (*entity.HouseObject, error) {
	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewTermQuery("house_guid", guid)).
		Size(1).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	var item *dto.JsonHouseDto
	// Конвертирует структуру ответа в DTO
	if len(res.Hits.Hits) > 0 {
		if err := json.Unmarshal(res.Hits.Hits[0].Source, &item); err != nil {
			return nil, err
		}

		return item.ToEntity(), nil
	}

	return nil, nil
}

// Получить дома по GUID
func (a *ElasticHouseRepository) GetByGuidList(guids []string) ([]*entity.HouseObject, error) {
	if len(guids) == 0 {
		return nil, nil
	}
	// Инициализирует сервис выборки элементов через ScrollApi
	scrollService := a.elasticClient.Client.Scroll(a.GetIndexName()).
		Query(elastic.NewTermsQuery("house_guid", util.ConvertStringSliceToInterface(guids)...))

	scrollData, err := a.elasticClient.ScrollData(scrollService, a.batchSize)
	if err != nil {
		a.logger.Error(err.Error())
	}

	var items []*entity.HouseObject
	var item *dto.JsonHouseDto

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

// Найти дома по GUID адресов
func (a *ElasticHouseRepository) GetByAddressGuidList(guids []string) ([]*entity.HouseObject, error) {
	if len(guids) == 0 {
		return nil, nil
	}

	// Инициализирует сервис выборки элементов через ScrollApi
	scrollService := a.elasticClient.Client.Scroll(a.GetIndexName()).
		Query(elastic.NewTermsQuery("ao_guid", util.ConvertStringSliceToInterface(guids)...)).
		Sort("house_full_num.keyword", true)

	return a.scroll(scrollService)
}

// Найти дома по GUID адреса
func (a *ElasticHouseRepository) GetByAddressGuid(guid string) ([]*entity.HouseObject, error) {
	if guid == "" {
		return nil, nil
	}

	return a.GetByAddressGuidList([]string{guid})
}

// Получить GUID последних обновленных домов
func (a *ElasticHouseRepository) GetLastUpdatedGuids(start time.Time) ([]string, error) {
	var guids []string

	// Инициализирует сервис выборки элементов через ScrollApi
	scrollService := a.elasticClient.Client.Scroll(a.GetIndexName()).
		Query(elastic.NewRangeQuery("bazis_update_date").Gte(start.Format(util.TimeFormat))).
		Scroll("10m")

	items, err := a.scroll(scrollService)

	if err != nil {
		return nil, err
	}
	for _, item := range items {
		guids = append(guids, item.AoGuid)
	}
	// Удаление дублей
	guids = util.UniqueStringSlice(guids)

	return guids, nil
}

// Найти дома по подстроке
func (a *ElasticHouseRepository) GetAddressByTerm(term string, size int64, from int64, filter ...entity.FilterObject) ([]*entity.HouseObject, error) {
	if size == 0 {
		size = 100
	}

	queries := []elastic.Query{elastic.NewMatchQuery("address_suggest", term).Operator("and")}
	queries = a.prepareFilter(queries, filter...)
	if queries == nil {
		return nil, nil
	}

	res, err := a.elasticClient.Client.
		Search(a.indexName).
		Query(elastic.NewBoolQuery().Must(queries...)).
		From(int(from)).
		Size(int(size)).
		Sort("full_address", true).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	var items []*entity.HouseObject
	var item *dto.JsonHouseDto
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
func (a *ElasticHouseRepository) prepareFilter(queries []elastic.Query, filters ...entity.FilterObject) []elastic.Query {
	for _, filter := range filters {
		if len(filter.Level.Values) > 0 {
			hasHouses := false
			for _, value := range filter.Level.Values {
				if value == 8 {
					hasHouses = true
					break
				}
			}
			if !hasHouses {
				return nil
			}
		}
		if filter.Level.Min > 0 || filter.Level.Max > 0 {
			if filter.Level.Min > 0 && filter.Level.Min > 8 {
				return nil
			}
			if filter.Level.Max > 0 && filter.Level.Max < 8 {
				return nil
			}
		}
		if len(filter.ParentGuid.Values) > 0 {
			queries = append(queries, elastic.NewTermsQuery("ao_guid", util.ConvertStringSliceToInterface(filter.ParentGuid.Values)...))
		}
		if len(filter.KladrId.Values) > 0 {
			return nil
		}
	}

	return queries
}

// Обновить коллекцию домов
func (a *ElasticHouseRepository) InsertUpdateCollection(wg *sync.WaitGroup, channel <-chan interface{}, count chan<- int, isFull bool) {
	defer wg.Done()
	var total uint64
	begin := time.Now()
	step := 1
	var deleted []string
	updated := make(map[string]dto.JsonHouseDto)

	// Цикл получения объекта дома из канала
	for d := range channel {
		if d == nil {
			break
		}
		total++
		saveItem := dto.JsonHouseDto{}
		saveItem.GetFromEntity(d.(entity.HouseObject))
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
				a.logger.WithFields(interfaces.LoggerFields{"step": step, "count": total}).Info("Add houses to index")
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
		a.logger.WithFields(interfaces.LoggerFields{"step": step, "count": total}).Info("Add houses to index")
	}
	a.logger.WithFields(interfaces.LoggerFields{"count": total, "execTime": humanize.RelTime(begin, time.Now(), "", "")}).Info("House import execution time")
	a.Refresh()
	count <- int(total)
}

// Сохраняет данные в эластик
func (a *ElasticHouseRepository) update(updated map[string]dto.JsonHouseDto, deleted []string, isFull bool) {
	bulk := a.GetBulkService()
	ctx := context.Background()
	// Дополняет элементы полями из БД
	if len(updated) > 0 && !isFull {
		var updatedKeys []string
		for k := range updated {
			updatedKeys = append(updatedKeys, k)
		}

		items, _ := a.GetByGuidList(updatedKeys)
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
		a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Add houses bulk commit failed")
	}
	if res != nil && res.Errors {
		a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("Add houses bulk commit failed")
	}
}

// Подсчитать количество домов в БД по фильтру
func (a *ElasticHouseRepository) CountAllData(query interface{}) (int64, error) {
	if query == nil {
		query = elastic.NewBoolQuery()
	}
	return a.elasticClient.CountAllData(a.GetIndexName(), query.(elastic.Query))
}

// Обновить индекс
func (a *ElasticHouseRepository) Refresh() {
	a.elasticClient.RefreshIndexes([]string{a.GetIndexName()})
}

// Получить объект для работы с пачками элементов
func (a *ElasticHouseRepository) GetBulkService() *elastic.BulkService {
	return a.elasticClient.Client.Bulk().Index(a.GetIndexName())
}

// Индексация домов
func (a *ElasticHouseRepository) Index(start time.Time, indexChan <-chan entity.IndexObject, GetIndexObjects repository.GetIndexObjects) error {
	done := make(chan bool)
	// Создает канал для сохранения объектов в индекс
	a.results = make(chan dto.JsonHouseDto, a.noOfWorkers)
	// Обновляет индекс
	a.Refresh()
	var total int64
	// Ищет элементы по дате
	if indexChan == nil {
		query := a.prepareIndexQuery(start)
		total, _ = a.CountAllData(query)
		go a.getItemsByQuery(query, GetIndexObjects)
	}
	// Обновляет элементы в индексе
	go a.saveIndexItems(total, done)
	// Создает пул задач на обработку элементов
	if indexChan != nil {
		a.createWorkerPool(a.noOfWorkers, indexChan)
	}
	<-done
	// Обновляет индекс
	a.Refresh()

	return nil
}

// Получить дома из канала адресов
func (a *ElasticHouseRepository) getItemsByAddress(wg *sync.WaitGroup, indexChan <-chan entity.IndexObject) {
	defer wg.Done()
	indexObjectList := make(map[string]entity.IndexObject)
	for d := range indexChan {
		indexObjectList[d.AoGuid] = d
		if len(indexObjectList) >= a.batchSize {
			a.prepareIndexChanHouses(indexObjectList)
		}
	}
	if len(indexObjectList) > 0 {
		a.prepareIndexChanHouses(indexObjectList)
	}
}

// Подготовить дома для индексации из канала адресов
func (a *ElasticHouseRepository) prepareIndexChanHouses(indexObjectList map[string]entity.IndexObject) {
	var guids []string
	for k := range indexObjectList {
		guids = append(guids, k)
	}
	// Получает список домов по GUID адресов
	houses, err := a.GetByAddressGuidList(guids)
	if err != nil {
		a.logger.WithFields(interfaces.LoggerFields{"error": err, "ao_guids": guids}).Fatal("Get houses failed")
		for k := range indexObjectList {
			delete(indexObjectList, k)
		}
		return
	}
	for _, house := range houses {
		saveItem := dto.JsonHouseDto{}
		indexObject, ok := indexObjectList[house.AoGuid]
		if ok {
			// Конвертирует объект дома в DTO
			saveItem.GetFromEntity(*house)
			// Заполняет данные для поиска из элемента канала
			a.prepareItem(&saveItem, indexObject)
			a.results <- saveItem
		}
	}
	// Очищает список объектов
	for k := range indexObjectList {
		delete(indexObjectList, k)
	}
}

// Получить дома по фильтру
func (a *ElasticHouseRepository) getItemsByQuery(query elastic.Query, GetIndexObjects repository.GetIndexObjects) {
	batch := a.batchSize
	// Ограничивает размер пачки при поиске
	if batch > 10000 {
		batch = 10000
	}

	// Инициализирует сервис выборки элементов через ScrollApi
	scrollService := a.elasticClient.Client.Scroll(a.GetIndexName()).
		Query(query).
		Size(batch)

	ctx := context.Background()
	scrollService.Scroll("1m")
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
		var list []dto.JsonHouseDto
		var guids []string

		// Добавляет элементы в пул задач
		for _, hit := range res.Hits.Hits {
			var item dto.JsonHouseDto
			// Конвертирует структуру ответа в DTO
			if err := json.Unmarshal(hit.Source, &item); err != nil {
				a.logger.Fatal(err.Error())
			}
			guids = append(guids, item.AoGuid)
			list = append(list, item)
		}

		objectsList := GetIndexObjects(guids)
		for _, item := range list {
			object, ok := objectsList[item.AoGuid]
			if ok {
				a.prepareItem(&item, object)
				a.results <- item
			}
		}
	}

	// Принудительно закрывает сервис выборки элементов
	err := scrollService.Clear(ctx)
	if err != nil {
		a.logger.Error(err.Error())
	}
	a.logger.WithFields(interfaces.LoggerFields{"count": count}).Info("Houses update count")

	close(a.results)
}

// Подготовить дома перед записью
func (a *ElasticHouseRepository) prepareItem(item *dto.JsonHouseDto, object entity.IndexObject) {
	// Формирует информацию об адресе объекта
	suggest := "дом д. " + item.HouseNum
	if item.StructNum != "" {
		suggest += ", строение стр. " + item.StructNum
	}
	if item.BuildNum != "" {
		suggest += ", корпус кор. " + item.BuildNum
	}
	item.AddressSuggest = object.AddressSuggest + ", " + suggest
	item.FullAddress = object.FullAddress + ", " + item.HouseFullNum
	// Устанавливает время обновления объекта
	item.UpdateBazisDate()
}

// Подготовить фильтр для получения элементов
func (a *ElasticHouseRepository) prepareIndexQuery(start time.Time) elastic.Query {
	var query elastic.Query
	var queries []elastic.Query
	// Добавляет фильтр на ограничение выборки по дате начала импорта
	queries = append(queries, elastic.NewRangeQuery("bazis_update_date").Gte(start.Format(util.TimeFormat)))
	query = elastic.NewBoolQuery().Must(queries...)

	return query
}

// Создать пул задач на обработку элементов
func (a *ElasticHouseRepository) createWorkerPool(noOfWorkers int, indexChan <-chan entity.IndexObject) {
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		// Подготавливает элементы перед сохранением в индекс
		go a.getItemsByAddress(&wg, indexChan)
	}
	wg.Wait()
	close(a.results)
}

// Обновить элементы в индексе
func (a *ElasticHouseRepository) saveIndexItems(total int64, done chan bool) {
	// Получает объект для работы с пачками элементов
	bulk := a.GetBulkService()
	ctx := context.Background()
	begin := time.Now()
	// Инициализация прогресс-бара
	bar := util.StartNewProgress(int(total), "Indexing houses", false)

	for d := range a.results {
		// Добавляет объект в очередь на сохранение
		bulk.Add(elastic.NewBulkIndexRequest().Id(d.ID).Doc(d))
		bar.Increment()
		// Отправляет запросы в эластик при превышении размера пачки
		if bulk.NumberOfActions() >= a.batchSize {
			res, err := bulk.Do(ctx)
			if err != nil {
				a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("House index bulk commit failed")
				os.Exit(1)
			}
			if res.Errors {
				a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("House index bulk commit failed")
				os.Exit(1)
			}
		}
	}

	// Отправляет оставшиеся запросы в эластик
	if bulk.NumberOfActions() > 0 {
		res, err := bulk.Do(ctx)
		if err != nil {
			a.logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("House index bulk commit failed")
			os.Exit(1)
		}
		if res.Errors {
			a.logger.WithFields(interfaces.LoggerFields{"error": a.elasticClient.GetBulkError(res)}).Fatal("House index bulk commit failed")
			os.Exit(1)
		}
	}
	bar.Finish()
	a.logger.WithFields(interfaces.LoggerFields{"execTime": humanize.RelTime(begin, time.Now(), "", "")}).Info("House index execution time")
	done <- true
}
