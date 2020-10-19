package elastic

import (
	"context"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/olivere/elastic/v7"
	"io"
)

// Объект-обёртка клиента эластика
type Client struct {
	Client *elastic.Client // Клиент эластика
}

// Инициализация объекта
func NewElasticClient(configInterface interfaces.ConfigInterface, logger interfaces.LoggerInterface) *Client {
	scheme := configInterface.GetString("elastic.scheme")
	user := configInterface.GetString("elastic.username")
	pass := configInterface.GetString("elastic.password")

	if scheme == "" {
		scheme = "http"
	}
	// Инициализация свойств подключения к клиенту
	options := []elastic.ClientOptionFunc{
		elastic.SetURL(scheme + "://" + configInterface.GetString("elastic.host")),
		elastic.SetSniff(configInterface.GetBool("elastic.sniff")),
		elastic.SetGzip(configInterface.GetBool("elastic.gzip")),
		elastic.SetErrorLog(logger),
		//elastic.SetTraceLog(logger),
	}
	// Проверка авторизации
	if user != "" && pass != "" {
		options = append(options, elastic.SetBasicAuth(user, pass))
	}

	// Подключение к эластику
	client, err := elastic.NewClient(options...)

	if err != nil {
		panic(err)
	}

	return &Client{
		Client: client,
	}
}

// Проверка наличия индекса
func (e *Client) IndexExists(index string) (bool, error) {
	ctx := context.Background()
	exists, err := e.Client.IndexExists(index).Do(ctx)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// Удаление индекса
func (e *Client) DropIndex(index string) error {
	ctx := context.Background()
	exists, err := e.IndexExists(index)
	if err != nil {
		return err
	}
	if exists {
		_, err := e.Client.DeleteIndex(index).Do(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// Создание индекса
func (e *Client) CreateIndex(index string, body string) error {
	ctx := context.Background()
	exists, err := e.IndexExists(index)
	if err != nil {
		return err
	}
	if !exists {
		_, err := e.Client.CreateIndex(index).Body(body).Do(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Добавление обработчика
func (e *Client) CreatePreprocessor(pipelineId, pipeline string) error {
	ctx := context.Background()
	_, err := e.Client.IngestPutPipeline(pipelineId).BodyString(pipeline).Do(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Обновление индекса
func (e *Client) RefreshIndexes(indexes []string) {
	ctx := context.Background()
	e.Client.Refresh(indexes...).Do(ctx)
	e.Client.Flush(indexes...).Do(ctx)
	e.Client.Forcemerge(indexes...).Do(ctx)
	e.Client.ClearCache(indexes...).Do(ctx)
}

// Получить элементы из индекса через ScrollApi
func (e *Client) ScrollData(scrollService *elastic.ScrollService) ([]elastic.SearchHit, error) {
	ctx := context.Background()
	scrollService.Scroll("1s")
	var totals []elastic.SearchHit

	// Получает данные из эластика пачками
	for {
		res, err := scrollService.Do(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if res == nil || len(res.Hits.Hits) == 0 {
			break
		}
		for _, hit := range res.Hits.Hits {
			totals = append(totals, *hit)
		}
	}

	// Принудительно закрывает сервис выборки элементов
	err := scrollService.Clear(ctx)
	if err != nil {
		return nil, err
	}

	return totals, nil
}

// Подсчитать количество элементов в БД по фильтру
func (e *Client) CountAllData(index string, query elastic.Query) (int64, error) {
	cnt := e.Client.Count(index)
	if query != nil {
		cnt.Query(query)
	}
	res, err := cnt.Do(context.Background())
	if err != nil {
		return 0, err
	}

	return res, nil
}

// Получает список ошибок при работе с пачками
func (e *Client) GetBulkError(bulk *elastic.BulkResponse) *elastic.ErrorDetails {
	var errorDetail *elastic.ErrorDetails

Loop:
	for {
		for _, resItems := range bulk.Items {
			for _, resItem := range resItems {
				if resItem.Error != nil {
					errorDetail = resItem.Error
					break Loop
				}
			}
		}
		break
	}

	return errorDetail
}
